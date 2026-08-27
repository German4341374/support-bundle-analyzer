BEGIN;
DROP TABLE IF EXISTS audit_events, plugin_registry, bundle_comparisons, reports, redaction_findings,
    timeline_events, evidence, findings, analyzer_runs, artifacts, analysis_sessions, bundles,
    analyzers, artifact_types CASCADE;
COMMIT;
