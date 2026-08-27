package sanitize

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
)

type FileResult struct {
	Path         string         `json:"path"`
	SourceSHA256 string         `json:"sourceSha256"`
	OutputSHA256 string         `json:"outputSha256,omitempty"`
	Replacements map[string]int `json:"replacements,omitempty"`
	Status       string         `json:"status"`
	Reason       string         `json:"reason,omitempty"`
}

type Manifest struct {
	SchemaVersion string         `json:"schemaVersion"`
	Profile       string         `json:"profile"`
	Files         []FileResult   `json:"files"`
	Summary       map[string]int `json:"summary"`
}

func Workspace(source, destination, profile string) (Manifest, error) {
	if profile != "standard" && profile != "strict" {
		return Manifest{}, fmt.Errorf("unsupported redaction profile %q", profile)
	}
	if _, err := os.Stat(destination); err == nil {
		return Manifest{}, fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return Manifest{}, err
	}
	sourceArtifacts := filepath.Join(source, "artifacts")
	if info, err := os.Stat(sourceArtifacts); err != nil || !info.IsDir() {
		return Manifest{}, fmt.Errorf("workspace artifacts directory is unavailable: %w", err)
	}
	staging := destination + ".partial"
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(filepath.Join(staging, "artifacts"), 0o700); err != nil {
		return Manifest{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(staging)
		}
	}()
	detector := redact.NewDetector()
	manifest := Manifest{SchemaVersion: model.SchemaVersion, Profile: profile, Summary: make(map[string]int)}
	err := filepath.WalkDir(sourceArtifacts, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(sourceArtifacts, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		sourceHash, err := fileHash(path)
		if err != nil {
			return err
		}
		result := FileResult{Path: relative, SourceSHA256: sourceHash}
		isText, err := textFile(path)
		if err != nil {
			return err
		}
		if !isText {
			result.Status = "excluded"
			result.Reason = "binary artifacts are excluded from sanitized exports"
			manifest.Files = append(manifest.Files, result)
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		redacted, counts := detector.Redact(string(content), profile)
		target := filepath.Join(staging, "artifacts", filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(redacted), 0o600); err != nil {
			return err
		}
		result.OutputSHA256, err = fileHash(target)
		if err != nil {
			return err
		}
		result.Status = "sanitized"
		result.Replacements = counts
		for kind, count := range counts {
			manifest.Summary[kind] += count
		}
		manifest.Files = append(manifest.Files, result)
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
	serialized, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Manifest{}, err
	}
	serialized = append(serialized, '\n')
	if err := os.WriteFile(filepath.Join(staging, "redaction-manifest.json"), serialized, 0o600); err != nil {
		return Manifest{}, err
	}
	if err := os.Rename(staging, destination); err != nil {
		return Manifest{}, err
	}
	committed = true
	return manifest, nil
}

func textFile(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()
	reader := bufio.NewReader(io.LimitReader(f, 8192))
	data, err := io.ReadAll(reader)
	if err != nil {
		return false, err
	}
	return !strings.ContainsRune(string(data), '\x00'), nil
}

func fileHash(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
