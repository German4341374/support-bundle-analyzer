package pipeline

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/German4341374/support-bundle-analyzer/internal/analyze"
	"github.com/German4341374/support-bundle-analyzer/internal/archiveutil"
	"github.com/German4341374/support-bundle-analyzer/internal/classifier"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
	"github.com/German4341374/support-bundle-analyzer/internal/report"
	"github.com/German4341374/support-bundle-analyzer/internal/workspace"
)

type Progress func(stage, message string)

type Options struct {
	Input    string
	Output   string
	Timezone string
	Limits   model.Limits
	Progress Progress
}

func Run(ctx context.Context, options Options) (model.Workspace, error) {
	if options.Limits.MaxFiles == 0 {
		options.Limits = model.DefaultLimits()
	}
	if options.Timezone == "" {
		options.Timezone = "UTC"
	}
	location, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("unknown IANA timezone %q: %w", options.Timezone, err)
	}
	if options.Output == "" {
		options.Output = "analysis-workspace"
	}
	if _, err := os.Stat(options.Output); err == nil {
		return model.Workspace{}, fmt.Errorf("output path already exists: %s", options.Output)
	} else if !os.IsNotExist(err) {
		return model.Workspace{}, err
	}
	analysisID, err := randomID()
	if err != nil {
		return model.Workspace{}, err
	}
	staging := options.Output + ".partial-" + analysisID
	if err := workspace.Prepare(staging); err != nil {
		return model.Workspace{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(staging)
		}
	}()
	started := time.Now().UTC()
	progress(options.Progress, "archive.inspection", "Inspecting archive structure and configured limits")
	if err := ctx.Err(); err != nil {
		return model.Workspace{}, err
	}
	progress(options.Progress, "archive.extraction", "Extracting archive into a private workspace")
	extracted, err := archiveutil.Extract(options.Input, filepath.Join(staging, "artifacts"), options.Limits)
	if err != nil {
		return model.Workspace{}, err
	}
	progress(options.Progress, "artifact.classification", "Hashing and classifying artifacts")
	artifacts, err := classifier.ClassifyTree(filepath.Join(staging, "artifacts"))
	if err != nil {
		return model.Workspace{}, err
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	detector := redact.NewDetector()
	result := model.Workspace{Root: staging}
	result.Manifest = model.Manifest{
		ToolVersion: model.ToolVersion, SchemaVersion: model.SchemaVersion, AnalysisID: analysisID,
		Started: started, InputSHA256: extracted.InputSHA256, InputName: filepath.Base(options.Input),
		ArtifactCount: len(artifacts), Artifacts: artifacts, Limits: options.Limits,
	}
	progress(options.Progress, "analyzer.started", "Running built-in log and HAR analyzers")
	logStart := time.Now()
	logWarnings := []string{}
	logFiles := 0
	harFiles := 0
	for _, artifact := range artifacts {
		if err := ctx.Err(); err != nil {
			return model.Workspace{}, err
		}
		absolute := filepath.Join(staging, "artifacts", filepath.FromSlash(artifact.Path))
		switch artifact.Type {
		case "generic-log", "json-log", "syslog", "nginx-access", "nginx-error", "apache-access", "apache-error", "php-log", "docker-log":
			logFiles++
			analysis, err := analyze.AnalyzeLog(absolute, artifact, location, options.Limits, detector)
			if err != nil {
				logWarnings = append(logWarnings, fmt.Sprintf("%s: %v", artifact.Path, err))
				continue
			}
			result.Findings = append(result.Findings, analysis.Findings...)
			result.Timeline = append(result.Timeline, analysis.Timeline...)
			result.Sensitive = append(result.Sensitive, analysis.Sensitive...)
			logWarnings = append(logWarnings, analysis.Warnings...)
		case "har":
			harFiles++
			analysis, err := analyze.AnalyzeHAR(absolute, artifact, detector)
			if err != nil {
				logWarnings = append(logWarnings, fmt.Sprintf("%s: %v", artifact.Path, err))
				continue
			}
			result.Findings = append(result.Findings, analysis.Findings...)
			result.Timeline = append(result.Timeline, analysis.Timeline...)
			result.Sensitive = append(result.Sensitive, analysis.Sensitive...)
		case "archive":
			result.Manifest.Warnings = append(result.Manifest.Warnings, fmt.Sprintf("nested archive %s was indexed but not extracted", artifact.Path))
		default:
			matches, scanErr := scanTextPrivacy(absolute, artifact.Path, detector)
			if scanErr == nil {
				result.Sensitive = append(result.Sensitive, matches...)
			}
		}
	}
	status := "completed"
	if len(logWarnings) > 0 {
		status = "completed_with_warnings"
	}
	result.Manifest.Analyzers = append(result.Manifest.Analyzers, model.AnalyzerRun{
		Name: "builtin-log-har", Version: model.ToolVersion, Status: status,
		DurationMS: time.Since(logStart).Milliseconds(), Warnings: logWarnings,
	})
	if logFiles == 0 {
		result.Manifest.Warnings = append(result.Manifest.Warnings, "no supported log artifacts were detected")
	}
	if harFiles == 0 {
		result.Manifest.Warnings = append(result.Manifest.Warnings, "no HAR artifacts were detected")
	}
	if len(result.Findings) > options.Limits.MaxFindings {
		result.Findings = result.Findings[:options.Limits.MaxFindings]
		result.Manifest.Warnings = append(result.Manifest.Warnings, "finding limit reached; additional findings were omitted")
	}
	progress(options.Progress, "timeline.building", "Sorting the unified timeline")
	result.Manifest.Completed = time.Now().UTC()
	result.Manifest.Warnings = append(result.Manifest.Warnings, logWarnings...)
	if err := workspace.Write(staging, result); err != nil {
		return model.Workspace{}, err
	}
	progress(options.Progress, "report.generating", "Building the offline static report")
	if err := report.Generate(filepath.Join(staging, "report"), report.Data{Manifest: result.Manifest, Findings: result.Findings, Timeline: result.Timeline, Sensitive: result.Sensitive}); err != nil {
		return model.Workspace{}, err
	}
	if err := os.Rename(staging, options.Output); err != nil {
		return model.Workspace{}, err
	}
	committed = true
	result.Root = options.Output
	progress(options.Progress, "analysis.completed", fmt.Sprintf("Analysis completed with %d artifact(s) and %d finding(s)", len(artifacts), len(result.Findings)))
	return result, nil
}

func scanTextPrivacy(filename, artifact string, detector *redact.Detector) ([]model.SensitiveMatch, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	counts := make(map[string]int)
	for scanner.Scan() {
		if strings.IndexByte(scanner.Text(), 0) >= 0 {
			return nil, nil
		}
		for _, match := range detector.Detect(scanner.Text(), true) {
			counts[match.Kind] += match.Count
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]model.SensitiveMatch, 0, len(keys))
	for _, key := range keys {
		result = append(result, model.SensitiveMatch{Artifact: artifact, Kind: key, Count: counts[key]})
	}
	return result, nil
}

func randomID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	buffer[6] = (buffer[6] & 0x0f) | 0x40
	buffer[8] = (buffer[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(buffer)
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}

func progress(handler Progress, stage, message string) {
	if handler != nil {
		handler(stage, message)
	}
}
