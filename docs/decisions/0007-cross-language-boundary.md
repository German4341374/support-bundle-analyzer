# ADR 0007: Keep cross-language models schema-first

Status: Accepted

Versioned JSON Schemas define artifacts, findings, timeline events, and plugin metadata. Every implementation validates its own output in contract tests. No language-specific serialization type is authoritative.

