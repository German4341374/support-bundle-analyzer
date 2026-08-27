# ADR 0004: Use PostgreSQL only for persistent server mode

Status: Accepted

Portable workspaces remain the source format for local use. PostgreSQL adds retention, concurrent sessions, cursor pagination, audit events, and full-text search for self-hosted server mode. Elasticsearch and a separate queue are not justified for the first release.

