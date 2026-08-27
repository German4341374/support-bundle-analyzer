#!/usr/bin/env bash
set -euo pipefail

: "${DATABASE_URL:?Set DATABASE_URL to a disposable PostgreSQL database}"
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f database/migrations/001_initial.up.sql
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f database/seed/development.sql
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -Atc "SELECT count(*) >= 10 FROM artifact_types" | grep -qx t
