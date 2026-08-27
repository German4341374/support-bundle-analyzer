package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/redact"
)

func TestAnalyzeHARFindsErrorsLatencyAndPrivacy(t *testing.T) {
	t.Parallel()
	name := filepath.Join(t.TempDir(), "capture.har")
	content := `{"log":{"entries":[{"startedDateTime":"2026-08-01T14:31:10Z","time":2500,"request":{"method":"GET","url":"https://api.example.test/orders?access_token=synthetic-value","headers":[{"name":"Authorization","value":"Bearer synthetic-token-value-123456"}],"cookies":[]},"response":{"status":503,"statusText":"Unavailable","headers":[],"cookies":[],"bodySize":1024,"content":{"size":1024,"mimeType":"application/json","text":""}},"timings":{"blocked":1,"dns":2,"connect":3,"ssl":4,"wait":2400}}]}}`
	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := model.Artifact{Path: "network/capture.har", SHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	result, err := AnalyzeHAR(name, artifact, redact.NewDetector())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 2 || len(result.Timeline) != 1 || len(result.Sensitive) == 0 {
		t.Fatalf("unexpected HAR analysis: findings=%d timeline=%d sensitive=%d", len(result.Findings), len(result.Timeline), len(result.Sensitive))
	}
}
