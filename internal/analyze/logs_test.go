package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
)

func TestAnalyzeLogGroupsEvidenceAndBuildsTimeline(t *testing.T) {
	t.Parallel()
	name := filepath.Join(t.TempDir(), "api.log")
	content := "2026-08-01T14:31:08Z ERROR service=api request_id=req-1 database connection refused user=1457\n" +
		"2026-08-01T14:31:09Z ERROR service=api request_id=req-2 database connection refused user=9182\n"
	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := model.Artifact{Path: "logs/api.log", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	result, err := AnalyzeLog(name, artifact, time.UTC, model.DefaultLimits(), redact.NewDetector())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) < 1 || len(result.Timeline) != 2 {
		t.Fatalf("unexpected result: findings=%d timeline=%d", len(result.Findings), len(result.Timeline))
	}
	if result.Findings[0].Evidence[0].LineStart != 1 || result.Findings[0].Evidence[0].LineEnd != 2 {
		t.Fatalf("finding evidence is not grouped: %+v", result.Findings[0].Evidence)
	}
}

func BenchmarkAnalyzeLog1MiB(b *testing.B) {
	root := b.TempDir()
	name := filepath.Join(root, "benchmark.log")
	line := "2026-08-01T14:31:08Z ERROR service=api request_id=req-123 database connection refused user=1457\n"
	content := strings.Repeat(line, (1<<20)/len(line))
	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		b.Fatal(err)
	}
	artifact := model.Artifact{Path: "benchmark.log", SHA256: strings.Repeat("a", 64)}
	limits := model.DefaultLimits()
	for b.Loop() {
		if _, err := AnalyzeLog(name, artifact, time.UTC, limits, redact.NewDetector()); err != nil {
			b.Fatal(err)
		}
	}
}

func TestFingerprintNormalizesVolatileIdentifiers(t *testing.T) {
	t.Parallel()
	first := Fingerprint("User 1457 failed request 88213 from 192.0.2.10")
	second := Fingerprint("User 9182 failed request 12818 from 198.51.100.20")
	if first != second {
		t.Fatalf("fingerprints differ: %q != %q", first, second)
	}
	if again := Fingerprint(first); again != first {
		t.Fatalf("normalization is not idempotent: %q != %q", again, first)
	}
}
