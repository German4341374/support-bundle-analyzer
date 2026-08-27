# Plugin system

Plugins use newline-delimited JSON so the Go core can stream records without a language-specific RPC dependency. Each plugin declares `name`, `version`, `protocolVersion`, executable arguments, supported artifact types, capabilities, timeout, and maintainer in `plugin.json`.

The core validates metadata before execution. Executables are launched directly with an argument vector; no shell command is constructed. Input is a single protocol request on standard input. Standard output accepts only protocol records and is capped. Standard error is diagnostic text with a separate cap and must not contain artifact content by default.

Valid terminal states are `completed`, `completed_with_warnings`, `failed`, `timed_out`, and `skipped`. Invalid JSON, excessive output, a deadline, cancellation, or a non-zero exit records an analyzer failure without aborting the entire session.

Protocol schemas live under `packages/schemas/v1`. Backward-incompatible changes require a new protocol version and an ADR. See [plugin SDK](../../packages/plugin-sdk/README.md) for a minimal implementation.
