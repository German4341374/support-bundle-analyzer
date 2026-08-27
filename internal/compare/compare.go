package compare

import (
	"sort"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

func Workspaces(baseline, incident model.Workspace) model.Comparison {
	result := model.Comparison{
		BaselineAnalysis:  baseline.Manifest.AnalysisID,
		IncidentAnalysis:  incident.Manifest.AnalysisID,
		BaselineArtifacts: len(baseline.Manifest.Artifacts),
		IncidentArtifacts: len(incident.Manifest.Artifacts),
		BaselineFindings:  len(baseline.Findings),
		IncidentFindings:  len(incident.Findings),
		BaselineEvents:    len(baseline.Timeline),
		IncidentEvents:    len(incident.Timeline),
		SeverityDelta:     make(map[string]int),
		AddedArtifacts:    []string{},
		RemovedArtifacts:  []string{},
		ChangedArtifacts:  []model.ArtifactChange{},
		NewFindings:       []model.Finding{},
		ResolvedFindings:  []model.Finding{},
		UnchangedRules:    []string{},
	}
	for _, finding := range baseline.Findings {
		result.SeverityDelta[finding.Severity]--
	}
	for _, finding := range incident.Findings {
		result.SeverityDelta[finding.Severity]++
	}
	baselineArtifacts := make(map[string]model.Artifact, len(baseline.Manifest.Artifacts))
	incidentArtifacts := make(map[string]model.Artifact, len(incident.Manifest.Artifacts))
	for _, artifact := range baseline.Manifest.Artifacts {
		baselineArtifacts[artifact.Path] = artifact
	}
	for _, artifact := range incident.Manifest.Artifacts {
		incidentArtifacts[artifact.Path] = artifact
		baselineArtifact, exists := baselineArtifacts[artifact.Path]
		if !exists {
			result.AddedArtifacts = append(result.AddedArtifacts, artifact.Path)
		} else if baselineArtifact.SHA256 != artifact.SHA256 {
			result.ChangedArtifacts = append(result.ChangedArtifacts, model.ArtifactChange{
				Path: artifact.Path, BaselineSHA256: baselineArtifact.SHA256, IncidentSHA256: artifact.SHA256,
			})
		}
	}
	for _, artifact := range baseline.Manifest.Artifacts {
		if _, exists := incidentArtifacts[artifact.Path]; !exists {
			result.RemovedArtifacts = append(result.RemovedArtifacts, artifact.Path)
		}
	}
	baselineByKey := make(map[string]model.Finding)
	incidentByKey := make(map[string]model.Finding)
	for _, finding := range baseline.Findings {
		baselineByKey[key(finding)] = finding
	}
	for _, finding := range incident.Findings {
		incidentByKey[key(finding)] = finding
	}
	for id, finding := range incidentByKey {
		if _, ok := baselineByKey[id]; !ok {
			result.NewFindings = append(result.NewFindings, finding)
		} else {
			result.UnchangedRules = append(result.UnchangedRules, finding.RuleID)
		}
	}
	for id, finding := range baselineByKey {
		if _, ok := incidentByKey[id]; !ok {
			result.ResolvedFindings = append(result.ResolvedFindings, finding)
		}
	}
	sort.Slice(result.NewFindings, func(i, j int) bool { return result.NewFindings[i].ID < result.NewFindings[j].ID })
	sort.Slice(result.ResolvedFindings, func(i, j int) bool { return result.ResolvedFindings[i].ID < result.ResolvedFindings[j].ID })
	sort.Strings(result.AddedArtifacts)
	sort.Strings(result.RemovedArtifacts)
	sort.Slice(result.ChangedArtifacts, func(i, j int) bool { return result.ChangedArtifacts[i].Path < result.ChangedArtifacts[j].Path })
	sort.Strings(result.UnchangedRules)
	result.UnchangedRules = unique(result.UnchangedRules)
	return result
}

func key(finding model.Finding) string {
	return finding.RuleID + "\x00" + finding.Component + "\x00" + finding.Title
}

func unique(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := []string{values[0]}
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
