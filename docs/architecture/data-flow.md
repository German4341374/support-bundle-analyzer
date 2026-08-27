# Analysis data flow

1. The CLI hashes the immutable source archive and records the configured limits.
2. Secure extraction normalizes every path before writing into a private staging directory.
3. The classifier combines filename, extension, magic bytes, directory context, and bounded content sniffing.
4. Built-in analyzers stream supported text and HAR artifacts and emit evidence-backed findings.
5. Optional JSONL plugins receive only the artifact and context required for their capability.
6. Events are normalized to UTC where the source timestamp is trustworthy; ambiguous timestamps retain warnings.
7. Findings, evidence, analyzer runs, privacy matches, and timeline events are written to versioned JSON/JSONL files.
8. The staging directory is promoted to a completed workspace only after its manifest is finalized.
9. The static report encodes normalized records and never embeds raw active HTML from an artifact.
10. Redaction creates a new export; it does not mutate the original evidence workspace.

Backpressure comes from bounded readers, worker limits, maximum findings, plugin stdout/stderr limits, and archive quotas. Failure of one artifact or plugin is recorded as a warning when the remaining evidence can still be analyzed safely.
