package workspace

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

func Prepare(root string) error {
	for _, directory := range []string{"artifacts", "normalized", "report", "metadata"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			return err
		}
	}
	return nil
}

func Write(root string, data model.Workspace) error {
	sort.Slice(data.Manifest.Artifacts, func(i, j int) bool { return data.Manifest.Artifacts[i].Path < data.Manifest.Artifacts[j].Path })
	sort.SliceStable(data.Findings, func(i, j int) bool {
		if data.Findings[i].Severity == data.Findings[j].Severity {
			return data.Findings[i].ID < data.Findings[j].ID
		}
		return severityRank(data.Findings[i].Severity) > severityRank(data.Findings[j].Severity)
	})
	sort.SliceStable(data.Timeline, func(i, j int) bool {
		if data.Timeline[i].Timestamp.Equal(data.Timeline[j].Timestamp) {
			return data.Timeline[i].Artifact < data.Timeline[j].Artifact
		}
		return data.Timeline[i].Timestamp.Before(data.Timeline[j].Timestamp)
	})
	if err := writeJSON(filepath.Join(root, "manifest.json"), data.Manifest); err != nil {
		return err
	}
	if err := writeJSONL(filepath.Join(root, "findings.jsonl"), data.Findings); err != nil {
		return err
	}
	if err := writeJSONL(filepath.Join(root, "timeline.jsonl"), data.Timeline); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "redaction.json"), map[string]any{"profile": "review-only", "findings": data.Sensitive}); err != nil {
		return err
	}
	return writeJSON(filepath.Join(root, "analyzer-runs.json"), data.Manifest.Analyzers)
}

func Load(root string) (model.Workspace, error) {
	var result model.Workspace
	result.Root = root
	if err := readJSON(filepath.Join(root, "manifest.json"), &result.Manifest); err != nil {
		return result, err
	}
	if err := readJSONL(filepath.Join(root, "findings.jsonl"), &result.Findings); err != nil {
		return result, err
	}
	if err := readJSONL(filepath.Join(root, "timeline.jsonl"), &result.Timeline); err != nil {
		return result, err
	}
	var redactionData struct {
		Findings []model.SensitiveMatch `json:"findings"`
	}
	if err := readJSON(filepath.Join(root, "redaction.json"), &redactionData); err != nil {
		return result, err
	}
	result.Sensitive = redactionData.Findings
	return result, nil
}

func writeJSON(name string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(name, content, 0o600)
}

func readJSON(name string, target any) error {
	content, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, target)
}

func writeJSONL[T any](name string, values []T) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(f)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			f.Close()
			return err
		}
	}
	return f.Close()
}

func readJSONL[T any](name string, values *[]T) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	for scanner.Scan() {
		var value T
		if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
			return err
		}
		*values = append(*values, value)
	}
	return scanner.Err()
}

func severityRank(value string) int {
	switch value {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	default:
		return 1
	}
}
