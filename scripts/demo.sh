#!/usr/bin/env bash
set -euo pipefail

binary="${SBA_BINARY:-./bin/support-bundle-analyzer}"
root="${SBA_DEMO_ROOT:-.tmp/demo}"
rm -rf "${root}"
mkdir -p "${root}"

"${binary}" generate-demo "${root}/database-outage.zip"
"${binary}" generate-demo "${root}/healthy.zip" --scenario healthy
"${binary}" analyze "${root}/healthy.zip" --output "${root}/healthy-analysis" --timezone UTC
"${binary}" analyze "${root}/database-outage.zip" --output "${root}/outage-analysis" --timezone UTC
"${binary}" diff "${root}/healthy-analysis" "${root}/outage-analysis" --output "${root}/comparison.json"
"${binary}" redact "${root}/outage-analysis" --profile strict --output "${root}/sanitized"

echo "Open ${root}/outage-analysis/report/index.html in a browser."
echo "Review the measured healthy-versus-outage delta in ${root}/comparison.json."
