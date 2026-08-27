# Five-minute demonstration

## Before the call

Build the binary and keep the repository open. Use only the included synthetic data.

```bash
make build
```

## 0:00–1:00 — frame the problem

Explain that support bundles mix untrusted archives, logs, HAR files and secrets. The system is local-first: it never uploads the bundle and never executes its contents.

## 1:00–2:00 — run the vertical slice

```bash
./bin/support-bundle-analyzer generate-demo database-outage.zip
./bin/support-bundle-analyzer analyze database-outage.zip --output investigation --timezone UTC
```

Point out staged progress, the immutable manifest, SHA-256 inventory and refusal to overwrite an existing workspace.

## 2:00–3:15 — inspect evidence

Open `investigation/report/index.html`. Filter findings to `high`, open the database connection finding, and show its artifact and line range. Move to the timeline and show the HAR 503 aligned with application log events. Stress that the report says “evidence,” not “root cause.”

## 3:15–4:00 — privacy and adversarial safety

```bash
./bin/support-bundle-analyzer redact investigation --profile strict --output sanitized-investigation
go test ./internal/archiveutil -run 'TestExtract'
```

Show the redaction manifest and explain why binary artifacts are excluded. Mention traversal, links, duplicates, size and compression-ratio controls.

## 4:00–5:00 — engineering depth

Show the JSONL plugin contract, polyglot analyzer directories, cursor-paginated Fastify API, PostgreSQL schema, CI security workflow, Docker non-root image and Helm security context. End with the honest limitations in `docs/project-status.md` and the roadmap to stable plugin orchestration.
