package compare

import (
	"testing"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

func TestDiffIdentityIsEmpty(t *testing.T) {
	t.Parallel()
	workspace := model.Workspace{Manifest: model.Manifest{AnalysisID: "same"}, Findings: []model.Finding{{ID: "1", RuleID: "RULE", Title: "Observed", Component: "api"}}}
	result := Workspaces(workspace, workspace)
	if len(result.NewFindings) != 0 || len(result.ResolvedFindings) != 0 {
		t.Fatalf("diff(A, A) is not empty: %+v", result)
	}
}
