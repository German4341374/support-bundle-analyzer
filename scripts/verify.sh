#!/usr/bin/env bash
set -euo pipefail

test -z "$(gofmt -l apps internal)"
go vet ./apps/... ./internal/...
staticcheck ./apps/... ./internal/...
go test -race ./apps/... ./internal/...
go build -trimpath -o .tmp/support-bundle-analyzer ./apps/cli

npm run format:check
npm run lint
npm run typecheck
npm test
npm run build
npm audit --audit-level=high

python3 -m ruff check analyzers/log-intelligence-python
python3 -m ruff format --check analyzers/log-intelligence-python
python3 -m mypy analyzers/log-intelligence-python/src
python3 -m pytest analyzers/log-intelligence-python

dotnet format analyzers/windows-diagnostics-analyzer/src/WindowsDiagnosticsAnalyzer.csproj --verify-no-changes
dotnet format analyzers/windows-diagnostics-analyzer/tests/WindowsDiagnosticsAnalyzer.Tests.csproj --verify-no-changes
dotnet test analyzers/windows-diagnostics-analyzer/tests/WindowsDiagnosticsAnalyzer.Tests.csproj --configuration Release
mvn -B -f analyzers/jvm-diagnostics-analyzer/pom.xml verify
(cd analyzers/php-web-diagnostics-analyzer && composer audit --locked && composer check)

actionlint
python3 -m yamllint -c .yamllint.yml .github compose.yaml deploy/helm
docker compose config --quiet
helm lint deploy/helm/support-bundle-analyzer
terraform -chdir=deploy/terraform fmt -check -recursive
terraform -chdir=deploy/terraform init -backend=false
terraform -chdir=deploy/terraform validate
docker build --target runtime -t support-bundle-analyzer:verify .

rm -rf .tmp/verify-demo
mkdir -p .tmp/verify-demo
.tmp/support-bundle-analyzer generate-demo .tmp/verify-demo/outage.zip
.tmp/support-bundle-analyzer generate-demo .tmp/verify-demo/healthy.zip --scenario healthy
.tmp/support-bundle-analyzer analyze .tmp/verify-demo/outage.zip --output .tmp/verify-demo/outage --timezone UTC --quiet
.tmp/support-bundle-analyzer analyze .tmp/verify-demo/healthy.zip --output .tmp/verify-demo/healthy --timezone UTC --quiet
.tmp/support-bundle-analyzer diff .tmp/verify-demo/healthy .tmp/verify-demo/outage --output .tmp/verify-demo/comparison.json
.tmp/support-bundle-analyzer redact .tmp/verify-demo/outage --profile strict --output .tmp/verify-demo/sanitized
test -s .tmp/verify-demo/outage/report/index.html
test -s .tmp/verify-demo/comparison.json
test -s .tmp/verify-demo/sanitized/redaction-manifest.json

echo "Full verification completed successfully."
