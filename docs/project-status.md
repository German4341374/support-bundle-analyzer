# Project status

Version `0.1.0` is an engineering preview with a working local analysis vertical slice. It is not presented as a production release.

## Implemented and locally verified on 28 August 2026

- Go secure extraction, classification, hashing, built-in log/HAR analysis, timeline, static report, strict redaction and comparison.
- Adversarial ZIP/TAR tests for traversal, absolute paths, symlinks, duplicate paths and compression-ratio limits.
- TypeScript API lint, strict typecheck, unit tests, build and dependency audit.
- Python log plugin Ruff, strict mypy and Pytest suite.
- .NET Windows analyzer Release build and xUnit suite.
- Java JVM analyzer Maven verify and shaded JAR build.
- PHP analyzer coding-style check, PHPStan at maximum level and PHPUnit suite.
- Deterministic healthy-versus-database-outage generation, artifact/finding/timeline comparison, strict sanitization, and six browser-verified screenshots.
- Fastify process deadlines, public minimal health/readiness probes, authenticated analysis routes, and terminal SSE completion.
- Helm lint/render, Terraform fmt/init/validate, Compose config, Hadolint, Actionlint, YAML lint, and Gitleaks source scan.

The exact local command log is in [verification.md](verification.md). This document never treats an unexecuted command as passing.

## Implemented but not locally runtime-tested in this Windows session

- Docker image and Compose stack (Docker Desktop is installed, but its Linux engine is unavailable because WSL2 is not active).
- PostgreSQL migration execution (no local PostgreSQL server was available).
- kind smoke test (it requires the unavailable local Linux container engine).
- GitHub Actions behavior, which can only be confirmed after publication and workflow completion.

## Known gaps

- The default Go pipeline does not yet auto-discover and invoke all bundled external plugins.
- The API's active-session store is in memory; the PostgreSQL schema is not wired through every API route.
- Full platform-specific Kubernetes and Docker analyzers, browser E2E, signed plugin manifests and SARIF export remain roadmap work.
- No release is created until the 1.0 Definition of Done is satisfied.
