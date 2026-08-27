package analyze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
)

var (
	uuidPattern        = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b`)
	ipPattern          = regexp.MustCompile(`\b(?:25[0-5]|2[0-4]\d|1?\d?\d)(?:\.(?:25[0-5]|2[0-4]\d|1?\d?\d)){3}\b`)
	numberPattern      = regexp.MustCompile(`\b\d{3,}\b`)
	correlationPattern = regexp.MustCompile(`(?i)(?:request[_-]?id|trace[_-]?id|correlation[_-]?id|x-request-id)[=:"\s]+([A-Za-z0-9._:-]+)`)
	volatileIDPattern  = regexp.MustCompile(`(?i)\b(request[_-]?id|trace[_-]?id|correlation[_-]?id|x-request-id)([=:"\s]+)[A-Za-z0-9._:-]+`)
	timestampPattern   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ][0-9:.]+(?:Z|[+-][0-9:]+)?`)
)

type logRule struct {
	ID          string
	Title       string
	Severity    string
	Category    string
	Pattern     *regexp.Regexp
	Explanation string
	NextSteps   []string
}

var logRules = []logRule{
	{ID: "LOG_CONNECTION_REFUSED", Title: "Repeated connection failures", Severity: "high", Category: "connectivity", Pattern: regexp.MustCompile(`(?i)connection (?:refused|reset)|ECONNREFUSED`), Explanation: "Evidence suggests that a component could not establish or retain a network connection. The destination availability and network path need verification.", NextSteps: []string{"Verify the destination service availability.", "Verify the configured host, port, and network path."}},
	{ID: "LOG_DATABASE_ERROR", Title: "Database errors observed", Severity: "high", Category: "database", Pattern: regexp.MustCompile(`(?i)(?:database|postgres|mysql|sql).*(?:error|failed|unavailable|timeout)|too many connections`), Explanation: "Database-related failures were observed in application output. This is evidence of a failed operation, not proof that the database itself is the root cause.", NextSteps: []string{"Check database health and connection capacity.", "Compare application and database timestamps."}},
	{ID: "LOG_AUTHENTICATION_FAILED", Title: "Authentication failures observed", Severity: "medium", Category: "authentication", Pattern: regexp.MustCompile(`(?i)authentication failed|invalid credentials|unauthorized|access denied|status[=: ]+401`), Explanation: "Authentication or authorization failures were recorded and may indicate invalid credentials, expired tokens, or policy denial.", NextSteps: []string{"Verify credential validity without exposing credentials.", "Inspect identity-provider and authorization-policy events."}},
	{ID: "LOG_TIMEOUT", Title: "Request or dependency timeouts observed", Severity: "medium", Category: "latency", Pattern: regexp.MustCompile(`(?i)timed? out|deadline exceeded|upstream timeout|ETIMEDOUT`), Explanation: "Operations exceeded a configured time boundary. Dependency latency and timeout configuration need verification.", NextSteps: []string{"Compare timeout values with observed dependency latency.", "Inspect events immediately before the timeout window."}},
	{ID: "LOG_OUT_OF_MEMORY", Title: "Out-of-memory evidence observed", Severity: "high", Category: "resources", Pattern: regexp.MustCompile(`(?i)outofmemoryerror|out of memory|oomkilled|cannot allocate memory`), Explanation: "Explicit memory exhaustion evidence was observed. Confirm process and host memory conditions before attributing a root cause.", NextSteps: []string{"Inspect memory limits and utilization.", "Check allocation pressure before the event."}},
	{ID: "LOG_DNS_FAILURE", Title: "DNS resolution failures observed", Severity: "medium", Category: "network", Pattern: regexp.MustCompile(`(?i)no such host|name or service not known|dns.*fail|NXDOMAIN`), Explanation: "Name-resolution failures were observed. DNS availability, search domains, and the requested hostname need verification.", NextSteps: []string{"Verify the hostname and DNS resolver configuration.", "Compare with network diagnostic artifacts."}},
}

type aggregate struct {
	rule        logRule
	count       int
	firstLine   int
	lastLine    int
	firstText   string
	component   string
	fingerprint string
}

type LogResult struct {
	Findings  []model.Finding
	Timeline  []model.TimelineEvent
	Sensitive []model.SensitiveMatch
	Warnings  []string
}

func AnalyzeLog(filename string, artifact model.Artifact, timezone *time.Location, limits model.Limits, detector *redact.Detector) (LogResult, error) {
	f, err := os.Open(filename)
	if err != nil {
		return LogResult{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	aggregates := make(map[string]*aggregate)
	result := LogResult{}
	lineNumber := 0
	privacyCounts := make(map[string]int)
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		parsed := parseLogLine(line, timezone)
		for _, match := range detector.Detect(line, true) {
			privacyCounts[match.Kind] += match.Count
		}
		for _, rule := range logRules {
			if !rule.Pattern.MatchString(line) {
				continue
			}
			fingerprint := Fingerprint(line)
			key := rule.ID + "\x00" + parsed.Component + "\x00" + fingerprint
			item := aggregates[key]
			if item == nil {
				item = &aggregate{rule: rule, firstLine: lineNumber, firstText: safeExcerpt(line), component: parsed.Component, fingerprint: fingerprint}
				aggregates[key] = item
			}
			item.count++
			item.lastLine = lineNumber
		}
		if parsed.Timestamp.IsZero() || len(result.Timeline) >= limits.MaxTimelineEvents || !isTimelineLevel(parsed.Severity) {
			continue
		}
		result.Timeline = append(result.Timeline, model.TimelineEvent{
			Timestamp: parsed.Timestamp, Source: "log", Component: parsed.Component, Severity: parsed.Severity,
			Category: "application-log", Message: safeExcerpt(parsed.Message), Artifact: artifact.Path,
			CorrelationID: parsed.CorrelationID, Evidence: map[string]any{"line": lineNumber},
		})
	}
	if err := scanner.Err(); err != nil {
		return LogResult{}, err
	}
	if len(result.Timeline) >= limits.MaxTimelineEvents {
		result.Warnings = append(result.Warnings, "timeline event limit reached; remaining events were not indexed")
	}
	keys := make([]string, 0, len(aggregates))
	for key := range aggregates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		item := aggregates[key]
		result.Findings = append(result.Findings, model.Finding{
			ID: stableID(artifact.SHA256, item.rule.ID, item.fingerprint), RuleID: item.rule.ID, Title: item.rule.Title,
			Severity: item.rule.Severity, Category: item.rule.Category, Component: item.component,
			Summary:     fmt.Sprintf("Observed %d matching event(s) in %s between lines %d and %d.", item.count, artifact.Path, item.firstLine, item.lastLine),
			Explanation: item.rule.Explanation, Confidence: "strong", Occurrences: item.count,
			Evidence:  []model.Evidence{{Artifact: artifact.Path, LineStart: item.firstLine, LineEnd: item.lastLine, Excerpt: item.firstText}},
			NextSteps: item.rule.NextSteps,
		})
	}
	privacyKinds := make([]string, 0, len(privacyCounts))
	for kind := range privacyCounts {
		privacyKinds = append(privacyKinds, kind)
	}
	sort.Strings(privacyKinds)
	for _, kind := range privacyKinds {
		result.Sensitive = append(result.Sensitive, model.SensitiveMatch{Artifact: artifact.Path, Kind: kind, Count: privacyCounts[kind]})
	}
	return result, nil
}

type parsedLog struct {
	Timestamp     time.Time
	Severity      string
	Component     string
	Message       string
	CorrelationID string
}

func parseLogLine(line string, timezone *time.Location) parsedLog {
	result := parsedLog{Severity: detectSeverity(line), Component: "unknown", Message: line}
	var object map[string]any
	if json.Unmarshal([]byte(line), &object) == nil {
		result.Timestamp = parseAnyTimestamp(firstString(object, "timestamp", "time", "ts", "@timestamp"), timezone)
		result.Severity = normalizeSeverity(firstString(object, "level", "severity", "log.level"))
		result.Component = firstString(object, "service", "component", "service.name", "logger")
		result.Message = firstString(object, "message", "msg", "error")
		result.CorrelationID = firstString(object, "request_id", "requestId", "trace_id", "traceId", "correlation_id", "correlationId", "x-request-id")
	} else {
		if raw := timestampPattern.FindString(line); raw != "" {
			result.Timestamp = parseAnyTimestamp(raw, timezone)
		}
		if match := correlationPattern.FindStringSubmatch(line); len(match) == 2 {
			result.CorrelationID = match[1]
		}
	}
	if result.Severity == "" {
		result.Severity = "info"
	}
	if result.Component == "" {
		result.Component = "unknown"
	}
	if result.Message == "" {
		result.Message = line
	}
	return result
}

func parseAnyTimestamp(value string, timezone *time.Location) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.000Z07:00", "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05.000", "2006-01-02 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, value, timezone); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func firstString(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := object[key]; ok {
			if text, ok := value.(string); ok {
				return text
			}
		}
	}
	return ""
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(value) {
	case "warn", "warning":
		return "warning"
	case "err", "error", "fatal":
		return "error"
	case "critical", "crit":
		return "critical"
	case "debug":
		return "debug"
	case "info", "information":
		return "info"
	default:
		return ""
	}
}

func detectSeverity(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "critical"), strings.Contains(lower, "fatal"):
		return "critical"
	case strings.Contains(lower, "error"), strings.Contains(lower, "failed"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warning"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}

func isTimelineLevel(value string) bool {
	return value == "warning" || value == "error" || value == "critical"
}

func Fingerprint(value string) string {
	normalized := timestampPattern.ReplaceAllString(value, "<TIMESTAMP>")
	normalized = volatileIDPattern.ReplaceAllString(normalized, "$1$2<ID>")
	normalized = uuidPattern.ReplaceAllString(normalized, "<UUID>")
	normalized = ipPattern.ReplaceAllString(normalized, "<IP>")
	normalized = numberPattern.ReplaceAllString(normalized, "<NUMBER>")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if len(normalized) > 240 {
		normalized = normalized[:240]
	}
	return normalized
}

func safeExcerpt(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 320 {
		return value[:320] + "…"
	}
	return value
}

func stableID(parts ...string) string {
	joined := strings.Join(parts, "\x00")
	var hash uint64 = 1469598103934665603
	for i := 0; i < len(joined); i++ {
		hash ^= uint64(joined[i])
		hash *= 1099511628211
	}
	return fmt.Sprintf("f-%016x", hash)
}
