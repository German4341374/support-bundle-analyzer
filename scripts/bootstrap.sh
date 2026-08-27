#!/usr/bin/env bash
set -euo pipefail

command -v go >/dev/null || { echo "Go 1.26+ is required." >&2; exit 1; }
command -v node >/dev/null || { echo "Node.js 24+ is required." >&2; exit 1; }
command -v python3 >/dev/null || { echo "Python 3.12+ is required." >&2; exit 1; }

go mod download
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
npm ci
python3 -m pip install -e "./analyzers/log-intelligence-python[dev]"

echo "Core development dependencies are ready. See CONTRIBUTING.md for polyglot and deployment prerequisites."
