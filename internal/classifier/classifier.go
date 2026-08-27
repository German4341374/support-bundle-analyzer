package classifier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

var (
	logLevelPattern = regexp.MustCompile(`(?i)\b(debug|info|warn(?:ing)?|error|critical|fatal)\b`)
	nginxAccess     = regexp.MustCompile(`^\S+\s+\S+\s+\S+\s+\[[^]]+\]\s+"[A-Z]+\s+\S+\s+HTTP/`)
)

func ClassifyTree(root string) ([]model.Artifact, error) {
	artifacts := make([]model.Artifact, 0)
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		artifact, err := ClassifyFile(name, relative)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, artifact)
		return nil
	})
	return artifacts, err
}

func ClassifyFile(name, relative string) (model.Artifact, error) {
	f, err := os.Open(name)
	if err != nil {
		return model.Artifact{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return model.Artifact{}, err
	}
	prefix := make([]byte, 64<<10)
	n, err := io.ReadFull(f, prefix)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return model.Artifact{}, err
	}
	prefix = prefix[:n]
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return model.Artifact{}, err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return model.Artifact{}, err
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	return model.Artifact{
		ID:     digest[:16],
		Path:   relative,
		Type:   Detect(relative, prefix),
		SHA256: digest,
		Size:   info.Size(),
	}, nil
}

func Detect(name string, prefix []byte) string {
	lower := strings.ToLower(filepath.ToSlash(name))
	base := filepath.Base(lower)
	trimmed := bytes.TrimSpace(prefix)
	text := string(trimmed)
	switch {
	case bytes.HasPrefix(prefix, []byte("PK\x03\x04")), bytes.HasPrefix(prefix, []byte("\x1f\x8b")):
		return "archive"
	case strings.HasSuffix(lower, ".hprof"):
		return "recognized-large-binary-artifact"
	case strings.HasSuffix(lower, ".har") || strings.Contains(text, `"log"`) && strings.Contains(text, `"entries"`) && strings.Contains(text, `"startedDateTime"`):
		return "har"
	case strings.HasSuffix(lower, ".evtx"):
		return "windows-event-binary"
	case strings.HasSuffix(lower, ".xml") && strings.Contains(text, "<Event") && strings.Contains(text, "System"):
		return "windows-event"
	case base == "docker-inspect.json" || strings.Contains(lower, "docker_inspect"):
		return "docker-inspect"
	case strings.Contains(lower, "gc.log") || strings.Contains(text, "[gc,") || strings.Contains(text, "Full GC"):
		return "jvm-gc-log"
	case strings.Contains(lower, "thread") && strings.Contains(text, "java.lang.Thread.State"):
		return "jvm-thread-dump"
	case strings.Contains(lower, "php") && strings.HasSuffix(lower, ".log"):
		return "php-log"
	case strings.Contains(lower, "nginx") && strings.Contains(lower, "access") || nginxAccess.MatchString(text):
		return "nginx-access"
	case strings.Contains(lower, "nginx") && strings.Contains(lower, "error"):
		return "nginx-error"
	case strings.Contains(lower, "apache") && strings.Contains(lower, "access"):
		return "apache-access"
	case strings.Contains(lower, "apache") && strings.Contains(lower, "error"):
		return "apache-error"
	case base == ".env" || strings.HasSuffix(lower, ".env"):
		return "env-config"
	case base == "package.json" || base == "composer.json" || base == "pom.xml" || base == "go.mod" || base == "requirements.txt" || base == "pyproject.toml" || strings.HasSuffix(base, ".csproj"):
		return "package-manifest"
	case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
		if strings.Contains(text, "apiVersion:") && strings.Contains(text, "kind:") {
			return "kubernetes-manifest"
		}
		return "yaml-config"
	case strings.HasSuffix(lower, ".toml"):
		return "toml-config"
	case strings.HasSuffix(lower, ".json") && jsonLogLike(text):
		return "json-log"
	case strings.HasSuffix(lower, ".json"):
		return "json-config"
	case networkDiagnosticName(base):
		return "network-diagnostic"
	case strings.HasSuffix(lower, ".log") && logLevelPattern.MatchString(text):
		return "generic-log"
	case strings.HasSuffix(lower, ".log") || strings.HasSuffix(lower, ".txt"):
		return "generic-log"
	case utf8.Valid(prefix) && !bytes.Contains(prefix, []byte{0}):
		return "unknown-text"
	default:
		return "unknown-binary"
	}
}

func jsonLogLike(text string) bool {
	return strings.Contains(text, `"timestamp"`) && (strings.Contains(text, `"level"`) || strings.Contains(text, `"severity"`))
}

func networkDiagnosticName(base string) bool {
	for _, marker := range []string{"ipconfig", "ifconfig", "ip-addr", "netstat", "traceroute", "tracert", "nslookup", "dig-output", "route"} {
		if strings.Contains(base, marker) {
			return true
		}
	}
	return false
}
