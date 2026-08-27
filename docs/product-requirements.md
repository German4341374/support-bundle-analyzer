# Product requirements

## Product statement

Universal Support Bundle Analyzer is an offline, privacy-first diagnostics workbench. It turns diagnostic archives into an evidence-indexed workspace instead of asking support engineers to inspect unrelated files manually.

## Primary user

Application support engineers, SREs, DevOps engineers, system administrators, and developers investigating customer-provided diagnostics.

## Core workflow

1. Inspect an archive without executing its contents.
2. Enforce extraction limits and reject unsafe entries.
3. Hash and classify artifacts.
4. Analyze generic logs and HAR files.
5. Emit evidence-backed findings and normalized timeline events.
6. detect sensitive values and build a privacy review.
7. Generate an autonomous static report.
8. Optionally compare two analyses or export a sanitized report.

## Product principles

- **Evidence first:** every finding cites an artifact location.
- **Deterministic first:** stable inputs and analyzer versions produce stable normalized output.
- **Privacy first:** no telemetry, cloud upload, remote processing, or external analysis is enabled.
- **Bounded processing:** archives, plugins, findings, and output streams have explicit limits.
- **Honest interpretation:** findings describe observations and possible contributing conditions, not guaranteed root causes.

## v0.1 acceptance criteria

- ZIP, TAR, TAR.GZ, TGZ, TAR.BZ2, TAR.XZ, and single-file GZIP ingestion.
- Safe rejection of path traversal, absolute paths, links, duplicate normalized paths, excessive compression, and configured size/count limits.
- Streaming log processing, HAR analysis, artifact hashes, findings, timeline, privacy review, and a network-free HTML report.
- Strict redaction with stable pseudonyms and no retained source secrets.
- Diff of healthy and incident workspaces.
- JSONL subprocess plugin contract with timeout and output limits.
- Synthetic database-outage demonstration and adversarial tests.

## Non-goals

This release is not a SIEM, antivirus, vulnerability scanner, cloud log collector, remote management platform, arbitrary-code execution environment, or automated root-cause oracle.

