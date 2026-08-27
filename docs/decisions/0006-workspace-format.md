# ADR 0006: Use append-friendly workspace files

Status: Accepted

Manifests use JSON while large findings and timeline collections use JSON Lines. This permits streaming, deterministic ordering, incremental checkpoints, and inspection with common tools.

