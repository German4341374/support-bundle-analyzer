package analyze

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/German4341374/support-bundle-analyzer/internal/apperror"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
)

type harDocument struct {
	Log struct {
		Entries []harEntry `json:"entries"`
	} `json:"log"`
}

type harEntry struct {
	StartedDateTime string  `json:"startedDateTime"`
	Time            float64 `json:"time"`
	Request         struct {
		Method   string      `json:"method"`
		URL      string      `json:"url"`
		Headers  []harHeader `json:"headers"`
		Cookies  []any       `json:"cookies"`
		PostData any         `json:"postData"`
	} `json:"request"`
	Response struct {
		Status     int         `json:"status"`
		StatusText string      `json:"statusText"`
		Headers    []harHeader `json:"headers"`
		Cookies    []any       `json:"cookies"`
		BodySize   int64       `json:"bodySize"`
		Content    struct {
			Size     int64  `json:"size"`
			MimeType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"content"`
	} `json:"response"`
	Timings struct {
		Blocked float64 `json:"blocked"`
		DNS     float64 `json:"dns"`
		Connect float64 `json:"connect"`
		SSL     float64 `json:"ssl"`
		Wait    float64 `json:"wait"`
	} `json:"timings"`
}

type harHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARResult struct {
	Findings  []model.Finding
	Timeline  []model.TimelineEvent
	Sensitive []model.SensitiveMatch
	Summary   map[string]any
}

func AnalyzeHAR(filename string, artifact model.Artifact, detector *redact.Detector) (HARResult, error) {
	f, err := os.Open(filename)
	if err != nil {
		return HARResult{}, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var document harDocument
	if err := decoder.Decode(&document); err != nil {
		return HARResult{}, apperror.Wrap(apperror.ArtifactMalformed, "HAR file is malformed", err)
	}
	result := HARResult{Summary: map[string]any{"requests": len(document.Log.Entries)}}
	statusCounts := map[string]int{"4xx": 0, "5xx": 0, "redirects": 0}
	domains := make(map[string]int)
	privacy := make(map[string]int)
	slow := 0
	totalBytes := int64(0)
	for index, entry := range document.Log.Entries {
		parsedURL, _ := url.Parse(entry.Request.URL)
		if parsedURL != nil && parsedURL.Hostname() != "" {
			domains[parsedURL.Hostname()]++
		}
		if entry.Response.Status >= 500 {
			statusCounts["5xx"]++
		} else if entry.Response.Status >= 400 {
			statusCounts["4xx"]++
		} else if entry.Response.Status >= 300 {
			statusCounts["redirects"]++
		}
		if entry.Time >= 2_000 {
			slow++
		}
		if entry.Response.BodySize > 0 {
			totalBytes += entry.Response.BodySize
		} else if entry.Response.Content.Size > 0 {
			totalBytes += entry.Response.Content.Size
		}
		for _, header := range append(entry.Request.Headers, entry.Response.Headers...) {
			lower := strings.ToLower(header.Name)
			if lower == "authorization" || lower == "cookie" || lower == "set-cookie" || lower == "x-api-key" {
				privacy[lower]++
			}
			for _, match := range detector.Detect(header.Value, true) {
				privacy[match.Kind] += match.Count
			}
		}
		for _, match := range detector.Detect(entry.Request.URL, true) {
			privacy[match.Kind] += match.Count
		}
		if entry.Request.PostData != nil || len(entry.Request.Cookies) > 0 || len(entry.Response.Cookies) > 0 {
			privacy["request_or_cookie_body"]++
		}
		timestamp, _ := time.Parse(time.RFC3339Nano, entry.StartedDateTime)
		if !timestamp.IsZero() {
			severity := "info"
			if entry.Response.Status >= 500 {
				severity = "error"
			} else if entry.Response.Status >= 400 || entry.Time >= 2_000 {
				severity = "warning"
			}
			result.Timeline = append(result.Timeline, model.TimelineEvent{
				Timestamp: timestamp.UTC(), Source: "har", Component: hostname(parsedURL), Severity: severity,
				Category: "http-request", Message: fmt.Sprintf("%s %s returned %d in %.0f ms", entry.Request.Method, safeURL(parsedURL), entry.Response.Status, entry.Time),
				Artifact: artifact.Path, Evidence: map[string]any{"jsonPointer": fmt.Sprintf("/log/entries/%d", index), "status": entry.Response.Status, "durationMs": entry.Time, "ttfbMs": entry.Timings.Wait},
			})
		}
	}
	result.Summary["totalTransferredBytes"] = totalBytes
	result.Summary["http4xx"] = statusCounts["4xx"]
	result.Summary["http5xx"] = statusCounts["5xx"]
	result.Summary["redirects"] = statusCounts["redirects"]
	result.Summary["slowRequests"] = slow
	result.Summary["domains"] = domains
	if statusCounts["5xx"] > 0 {
		result.Findings = append(result.Findings, harFinding(artifact, "HAR_HTTP_5XX", "HTTP server errors observed", "high", statusCounts["5xx"], "server-error", "Evidence shows HTTP 5xx responses. Application and upstream events around the same timestamps need verification."))
	}
	if statusCounts["4xx"] > 0 {
		result.Findings = append(result.Findings, harFinding(artifact, "HAR_HTTP_4XX", "HTTP client or authorization errors observed", "medium", statusCounts["4xx"], "client-error", "Evidence shows HTTP 4xx responses. Request validity and authorization policy need verification."))
	}
	if slow > 0 {
		result.Findings = append(result.Findings, harFinding(artifact, "HAR_SLOW_REQUESTS", "Slow HTTP requests observed", "medium", slow, "latency", "Requests exceeded the built-in two-second review threshold. Timing phases and server-side events should be compared."))
	}
	privacyKinds := make([]string, 0, len(privacy))
	for kind := range privacy {
		privacyKinds = append(privacyKinds, kind)
	}
	sort.Strings(privacyKinds)
	for _, kind := range privacyKinds {
		result.Sensitive = append(result.Sensitive, model.SensitiveMatch{Artifact: artifact.Path, Kind: kind, Count: privacy[kind]})
	}
	return result, nil
}

func harFinding(artifact model.Artifact, rule, title, severity string, count int, category, explanation string) model.Finding {
	return model.Finding{
		ID: stableID(artifact.SHA256, rule), RuleID: rule, Title: title, Severity: severity, Category: category,
		Summary: fmt.Sprintf("Observed %d matching request(s) in %s.", count, artifact.Path), Explanation: explanation,
		Confidence: "exact", Occurrences: count,
		Evidence:  []model.Evidence{{Artifact: artifact.Path, JSONPointer: "/log/entries"}},
		NextSteps: []string{"Inspect the affected requests and timing phases.", "Correlate request timestamps with application and dependency logs."},
	}
}

func hostname(parsed *url.URL) string {
	if parsed == nil || parsed.Hostname() == "" {
		return "unknown"
	}
	return parsed.Hostname()
}

func safeURL(parsed *url.URL) string {
	if parsed == nil {
		return "invalid-url"
	}
	return parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
}
