# ADR 0005: Stable scoped pseudonyms for redaction

Status: Accepted

Redaction replaces sensitive values with deterministic labels scoped to one export, such as `EMAIL_001`. This preserves relationships without storing a reverse mapping or the original value. Standard and strict profiles are built in; custom policies are versioned configuration.

