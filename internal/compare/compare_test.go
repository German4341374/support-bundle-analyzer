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
	if result.NewFindings == nil || result.ResolvedFindings == nil || result.ChangedArtifacts == nil {
		t.Fatal("comparison collections must encode as arrays, not null")
	}
}

func TestDiffReportsArtifactAndSeverityChanges(t *testing.T) {
	t.Parallel()
	baseline := model.Workspace{
		Manifest: model.Manifest{AnalysisID: "baseline", Artifacts: []model.Artifact{{Path: "app.log", SHA256: "old"}}},
		Findings: []model.Finding{{ID: "old", RuleID: "OLD", Title: "Old", Severity: "low"}},
		Timeline: []model.TimelineEvent{{}},
	}
	incident := model.Workspace{
		Manifest: model.Manifest{AnalysisID: "incident", Artifacts: []model.Artifact{{Path: "app.log", SHA256: "new"}, {Path: "trace.har", SHA256: "har"}}},
		Findings: []model.Finding{{ID: "new", RuleID: "NEW", Title: "New", Severity: "high"}},
		Timeline: []model.TimelineEvent{{}, {}},
	}
	result := Workspaces(baseline, incident)
	if len(result.ChangedArtifacts) != 1 || len(result.AddedArtifacts) != 1 {
		t.Fatalf("unexpected artifact delta: %+v", result)
	}
	if result.SeverityDelta["high"] != 1 || result.SeverityDelta["low"] != -1 {
		t.Fatalf("unexpected severity delta: %+v", result.SeverityDelta)
	}
}
