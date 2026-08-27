package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

func TestGenerateDoesNotEmbedRawXSS(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	payload := `<script>globalThis.compromised=true</script>`
	data := Data{Manifest: model.Manifest{ArtifactCount: 1}, Findings: []model.Finding{{ID: "x", RuleID: "XSS_TEST", Title: payload, Severity: "info", Category: "test", Summary: payload, Confidence: "exact", Evidence: []model.Evidence{{Artifact: "x.log", Excerpt: payload}}, NextSteps: []string{"review"}}}}
	if err := Generate(directory, data); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", "data.js"} {
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), payload) {
			t.Fatalf("raw XSS payload appears in %s", name)
		}
	}
}
