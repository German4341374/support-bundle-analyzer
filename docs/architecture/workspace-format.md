# Workspace format

A completed workspace is an investigation record, not a cache directory.

```text
workspace/
├── manifest.json
├── artifacts/
├── normalized/
├── findings.jsonl
├── timeline.jsonl
├── redaction.json
├── analyzer-runs.json
├── report/
└── metadata/
```

`manifest.json` records schema/tool versions, analysis identifier, source SHA-256, start/completion times, artifact count, analyzers, warnings, and effective limits. Artifact paths are normalized relative paths; source archives and absolute host paths are never embedded in a shareable report.

The analyzer refuses to overwrite an existing output. It builds in a private staging directory and promotes the directory only after successful finalization. Consumers must treat unknown fields as forward-compatible and must reject unsupported major schema versions.

The original workspace is immutable by convention. Redaction and comparison create separate outputs with provenance back to workspace hashes. Deleting an analysis means deleting its workspace and any separately persisted server-mode rows according to the retention policy.
