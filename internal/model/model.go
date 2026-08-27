package model

import "time"

const (
	ToolVersion   = "0.1.0"
	SchemaVersion = "1"
)

type Limits struct {
	MaxFiles             int     `json:"maxFiles"`
	MaxTotalBytes        int64   `json:"maxTotalBytes"`
	MaxSingleFileBytes   int64   `json:"maxSingleFileBytes"`
	MaxNestedDepth       int     `json:"maxNestedDepth"`
	MaxCompressionRatio  float64 `json:"maxCompressionRatio"`
	MaxFilenameBytes     int     `json:"maxFilenameBytes"`
	MaxTimelineEvents    int     `json:"maxTimelineEvents"`
	MaxFindings          int     `json:"maxFindings"`
	PluginTimeoutSeconds int     `json:"pluginTimeoutSeconds"`
}

func DefaultLimits() Limits {
	return Limits{
		MaxFiles:             100_000,
		MaxTotalBytes:        5 << 30,
		MaxSingleFileBytes:   1 << 30,
		MaxNestedDepth:       1,
		MaxCompressionRatio:  250,
		MaxFilenameBytes:     512,
		MaxTimelineEvents:    100_000,
		MaxFindings:          10_000,
		PluginTimeoutSeconds: 120,
	}
}

type Artifact struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type Evidence struct {
	Artifact    string `json:"artifact"`
	LineStart   int    `json:"lineStart,omitempty"`
	LineEnd     int    `json:"lineEnd,omitempty"`
	JSONPointer string `json:"jsonPointer,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Excerpt     string `json:"excerpt,omitempty"`
}

type Finding struct {
	ID          string     `json:"id"`
	RuleID      string     `json:"ruleId"`
	Title       string     `json:"title"`
	Severity    string     `json:"severity"`
	Category    string     `json:"category"`
	Component   string     `json:"component,omitempty"`
	Summary     string     `json:"summary"`
	Explanation string     `json:"explanation,omitempty"`
	Confidence  string     `json:"confidence"`
	Evidence    []Evidence `json:"evidence"`
	NextSteps   []string   `json:"nextSteps"`
	References  []string   `json:"references,omitempty"`
	Occurrences int        `json:"occurrences,omitempty"`
}

type TimelineEvent struct {
	Timestamp     time.Time      `json:"timestamp"`
	Source        string         `json:"source"`
	Component     string         `json:"component"`
	Severity      string         `json:"severity"`
	Category      string         `json:"category"`
	Message       string         `json:"message"`
	Artifact      string         `json:"artifact"`
	CorrelationID string         `json:"correlationId,omitempty"`
	Evidence      map[string]any `json:"evidence"`
}

type SensitiveMatch struct {
	Artifact string `json:"artifact"`
	Line     int    `json:"line,omitempty"`
	Kind     string `json:"kind"`
	Count    int    `json:"count"`
}

type AnalyzerRun struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Status     string   `json:"status"`
	DurationMS int64    `json:"durationMs"`
	Warnings   []string `json:"warnings,omitempty"`
}

type Manifest struct {
	ToolVersion   string        `json:"toolVersion"`
	SchemaVersion string        `json:"schemaVersion"`
	AnalysisID    string        `json:"analysisId"`
	Started       time.Time     `json:"started"`
	Completed     time.Time     `json:"completed"`
	InputSHA256   string        `json:"inputSha256"`
	InputName     string        `json:"inputName"`
	ArtifactCount int           `json:"artifactCount"`
	Artifacts     []Artifact    `json:"artifacts"`
	Analyzers     []AnalyzerRun `json:"analyzers"`
	Warnings      []string      `json:"warnings"`
	Limits        Limits        `json:"limits"`
}

type Workspace struct {
	Root      string
	Manifest  Manifest
	Findings  []Finding
	Timeline  []TimelineEvent
	Sensitive []SensitiveMatch
}

type RedactionSummary struct {
	Profile        string         `json:"profile"`
	FilesProcessed int            `json:"filesProcessed"`
	Replacements   map[string]int `json:"replacements"`
}

type Comparison struct {
	BaselineAnalysis  string           `json:"baselineAnalysis"`
	IncidentAnalysis  string           `json:"incidentAnalysis"`
	BaselineArtifacts int              `json:"baselineArtifacts"`
	IncidentArtifacts int              `json:"incidentArtifacts"`
	BaselineFindings  int              `json:"baselineFindings"`
	IncidentFindings  int              `json:"incidentFindings"`
	BaselineEvents    int              `json:"baselineTimelineEvents"`
	IncidentEvents    int              `json:"incidentTimelineEvents"`
	SeverityDelta     map[string]int   `json:"severityDelta"`
	AddedArtifacts    []string         `json:"addedArtifacts"`
	RemovedArtifacts  []string         `json:"removedArtifacts"`
	ChangedArtifacts  []ArtifactChange `json:"changedArtifacts"`
	NewFindings       []Finding        `json:"newFindings"`
	ResolvedFindings  []Finding        `json:"resolvedFindings"`
	UnchangedRules    []string         `json:"unchangedRules"`
}

type ArtifactChange struct {
	Path           string `json:"path"`
	BaselineSHA256 string `json:"baselineSha256"`
	IncidentSHA256 string `json:"incidentSha256"`
}
