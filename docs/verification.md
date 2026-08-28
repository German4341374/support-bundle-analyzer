# Verification record

This record distinguishes commands that actually ran from checks delegated to GitHub Actions. Results are not implied from configuration files alone.

## Local verification — 28 August 2026

The following commands completed with exit code `0` on Windows 10:

```text
gofmt -l apps internal
go vet ./apps/... ./internal/...
staticcheck ./apps/... ./internal/...
go test ./apps/... ./internal/...
go build -trimpath -o .tmp/support-bundle-analyzer.exe ./apps/cli
npm run format:check
npm run lint
npm run typecheck
npm test
npm run build
npm audit --audit-level=high
python -m ruff check analyzers/log-intelligence-python
python -m ruff format --check analyzers/log-intelligence-python
python -m mypy analyzers/log-intelligence-python/src
python -m pytest analyzers/log-intelligence-python
dotnet format <both analyzer projects> --verify-no-changes
dotnet test analyzers/windows-diagnostics-analyzer/tests/WindowsDiagnosticsAnalyzer.Tests.csproj --configuration Release
mvn -B -f analyzers/jvm-diagnostics-analyzer/pom.xml verify
composer audit --locked
php-cs-fixer check --diff
phpstan analyse --level=max src tests
phpunit
actionlint
python -m yamllint -c .yamllint.yml .github compose.yaml deploy/helm
gitleaks detect --source . --no-git --config .gitleaks.toml --redact
hadolint Dockerfile
docker compose config --quiet
helm lint deploy/helm/support-bundle-analyzer
helm template sba deploy/helm/support-bundle-analyzer --set auth.existingSecret=synthetic-secret
terraform -chdir=deploy/terraform fmt -check -recursive
terraform -chdir=deploy/terraform init -backend=false -input=false
terraform -chdir=deploy/terraform validate
```

Test results: 21 Go tests passed and one symlink test was skipped on Windows; 12 Vitest tests, 11 Pytest tests, 3 xUnit tests, 3 JUnit tests, and 3 PHPUnit tests passed. Node coverage was 45.99% statements, 55.62% branches, 39.58% functions, and 47.03% lines. Dependency audits reported no known npm or Composer advisories at the time of execution.

The real CLI also generated both synthetic scenarios, analyzed both archives, produced four outage findings versus zero healthy findings, compared three changed artifacts, generated both offline reports, and created a strict sanitized export whose manifest recorded one bearer-token and one connection-string replacement. Browser verification rendered overview, findings, timeline, HAR inventory, privacy, and comparison views without console errors.

A separate `git clone --no-local` clean checkout successfully installed the Go, npm, and Python dependencies; ran the portable Go, TypeScript, and Python checks; built the CLI and API; generated the synthetic outage archive; analyzed it; and produced the offline report.

Initial validation exposed malformed YAML flow mappings in three issue forms and Gitleaks false positives inside ignored Composer dependencies. The forms were corrected, and `.gitleaks.toml` now excludes generated dependency/build directories while continuing to scan repository source and fixtures.

## Checks not completed locally

- `go test -race` requires a Windows C compiler for Go's race detector; the Linux CI job runs this gate.
- Docker Desktop is installed, but its Linux engine cannot start in this session because WSL2 is not available. Container build, runtime health, Compose health, Trivy, SBOM, and kind are therefore CI-owned checks.
- The PostgreSQL migration smoke test requires a running PostgreSQL service and is executed by the integration workflow.

See [project status](project-status.md) for feature limitations. A release remains blocked until all Definition of Done gates, including browser E2E and kind, are green.
