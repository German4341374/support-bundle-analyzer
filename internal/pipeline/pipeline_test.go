package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/German4341374/support-bundle-analyzer/internal/synthetic"
)

func TestDatabaseOutageVerticalSlice(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bundle := filepath.Join(root, "database-outage.zip")
	if err := synthetic.WriteBundle(bundle); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "workspace")
	result, err := Run(context.Background(), Options{Input: bundle, Output: output, Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Manifest.ArtifactCount != 4 || len(result.Findings) < 2 || len(result.Timeline) < 2 || len(result.Sensitive) == 0 {
		t.Fatalf("unexpected vertical slice result: artifacts=%d findings=%d timeline=%d sensitive=%d", result.Manifest.ArtifactCount, len(result.Findings), len(result.Timeline), len(result.Sensitive))
	}
	for _, name := range []string{"manifest.json", "findings.jsonl", "timeline.jsonl", filepath.Join("report", "index.html"), filepath.Join("report", "data.js")} {
		if _, err := os.Stat(filepath.Join(output, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}
