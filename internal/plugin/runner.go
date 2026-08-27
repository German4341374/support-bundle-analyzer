package plugin

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/German4341374/support-bundle-analyzer/internal/apperror"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

type Manifest struct {
	Name                   string   `json:"name"`
	Version                string   `json:"version"`
	ProtocolVersion        string   `json:"protocolVersion"`
	Executable             []string `json:"executable"`
	SupportedArtifactTypes []string `json:"supportedArtifactTypes"`
	Capabilities           []string `json:"capabilities"`
	Timeout                int      `json:"timeout"`
	Maintainer             string   `json:"maintainer"`
}

type Request struct {
	ProtocolVersion string `json:"protocolVersion"`
	AnalysisID      string `json:"analysisId"`
	Artifact        struct {
		Path   string `json:"path"`
		Type   string `json:"type"`
		SHA256 string `json:"sha256"`
	} `json:"artifact"`
	Context struct {
		Timezone string `json:"timezone"`
	} `json:"context"`
}

type Response struct {
	Type    string               `json:"type"`
	Finding *model.Finding       `json:"finding,omitempty"`
	Event   *model.TimelineEvent `json:"event,omitempty"`
	Warning string               `json:"warning,omitempty"`
}

type Result struct {
	Status   string
	Findings []model.Finding
	Events   []model.TimelineEvent
	Warnings []string
	Stderr   string
}

type Runner struct {
	MaxStdoutBytes int64
	MaxStderrBytes int64
	MaxFindings    int
}

func (r Runner) Run(ctx context.Context, manifest Manifest, request Request) (Result, error) {
	if len(manifest.Executable) == 0 || manifest.ProtocolVersion != "1" {
		return Result{Status: "failed"}, apperror.New(apperror.AnalyzerFailed, "plugin manifest is invalid or unsupported")
	}
	timeout := time.Duration(manifest.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(ctx, manifest.Executable[0], manifest.Executable[1:]...)
	stdin, err := command.StdinPipe()
	if err != nil {
		return Result{Status: "failed"}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return Result{Status: "failed"}, err
	}
	var stderr bytes.Buffer
	command.Stderr = &limitWriter{writer: &stderr, remaining: defaultLimit(r.MaxStderrBytes, 1<<20)}
	if err := command.Start(); err != nil {
		return Result{Status: "failed"}, apperror.Wrap(apperror.AnalyzerFailed, "cannot start analyzer", err)
	}
	encodeErr := json.NewEncoder(stdin).Encode(request)
	closeErr := stdin.Close()
	if encodeErr != nil || closeErr != nil {
		_ = command.Process.Kill()
		return Result{Status: "failed"}, errors.Join(encodeErr, closeErr)
	}
	result := Result{Status: "completed"}
	limited := io.LimitReader(stdout, defaultLimit(r.MaxStdoutBytes, 8<<20)+1)
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		var response Response
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			result.Status = "completed_with_warnings"
			result.Warnings = append(result.Warnings, "analyzer emitted invalid JSON")
			continue
		}
		switch response.Type {
		case "finding":
			if response.Finding != nil && len(result.Findings) < defaultInt(r.MaxFindings, 10_000) {
				result.Findings = append(result.Findings, *response.Finding)
			}
		case "timeline_event":
			if response.Event != nil {
				result.Events = append(result.Events, *response.Event)
			}
		case "warning":
			result.Warnings = append(result.Warnings, response.Warning)
		default:
			result.Warnings = append(result.Warnings, "analyzer emitted an unsupported response type")
		}
	}
	scanErr := scanner.Err()
	waitErr := command.Wait()
	result.Stderr = stderr.String()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Status = "timed_out"
		return result, apperror.New(apperror.AnalyzerTimeout, "analyzer exceeded its timeout")
	}
	if scanErr != nil {
		result.Status = "failed"
		return result, apperror.Wrap(apperror.AnalyzerFailed, "cannot read analyzer output", scanErr)
	}
	if waitErr != nil {
		result.Status = "failed"
		return result, apperror.Wrap(apperror.AnalyzerFailed, "analyzer process failed", waitErr)
	}
	if len(result.Warnings) > 0 {
		result.Status = "completed_with_warnings"
	}
	return result, nil
}

type limitWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *limitWriter) Write(p []byte) (int, error) {
	original := len(p)
	if w.remaining <= 0 {
		return original, nil
	}
	if int64(len(p)) > w.remaining {
		p = p[:w.remaining]
	}
	written, err := w.writer.Write(p)
	w.remaining -= int64(written)
	if err != nil {
		return written, fmt.Errorf("write bounded plugin stderr: %w", err)
	}
	return original, nil
}

func defaultLimit(value, fallback int64) int64 {
	if value > 0 {
		return value
	}
	return fallback
}

func defaultInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
