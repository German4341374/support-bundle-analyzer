# ADR 0003: Default to local-only processing

Status: Accepted

The default product opens no outbound connections, sends no telemetry, and binds servers only to `127.0.0.1`. Diagnostic bundles regularly contain credentials and proprietary context, so remote processing must be a separate explicit deployment decision.

