# Getting started

Support Bundle Analyzer turns an untrusted diagnostic archive into a private investigation workspace. It does not upload data or execute files from the bundle.

## Prerequisites

- Go 1.26 or newer
- Git
- Node.js 24 and Python 3.12 for API and plugin development
- WSL2 or Linux for the complete `make verify` workflow

## First analysis

```bash
go build -trimpath -o bin/support-bundle-analyzer ./apps/cli
./bin/support-bundle-analyzer generate-demo .tmp/database-outage.zip
./bin/support-bundle-analyzer analyze .tmp/database-outage.zip --output .tmp/investigation --timezone UTC
```

Open `.tmp/investigation/report/index.html`. Review `manifest.json`, `findings.jsonl`, `timeline.jsonl`, and `redaction.json` before drawing conclusions.

## Before sharing

```bash
./bin/support-bundle-analyzer redact .tmp/investigation --profile strict --output .tmp/sanitized
```

Inspect `.tmp/sanitized/redaction-manifest.json`. Redaction is risk reduction, not a guarantee. Never publish a real customer bundle or an unreviewed report.

## Next steps

- Read the [compatibility matrix](reference/compatibility.md).
- Learn the [workspace format](architecture/workspace-format.md).
- Review the [threat model](security/threat-model.md).
- Add analyzers using the [plugin development guide](guides/plugin-development.md).
