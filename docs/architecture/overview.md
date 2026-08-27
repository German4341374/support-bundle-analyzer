# Architecture overview

The trust boundary starts at the input archive. The Go core owns bounded extraction, path normalization, hashing, artifact classification, orchestration, and immutable workspace creation. It never imports or executes bundle content.

Language-specific analyzers communicate through protocol version 1 JSON Lines. A plugin receives metadata and a workspace-relative artifact path, then emits bounded findings or timeline events. The core starts plugins with argument arrays, a deadline, cancellation, and output limits. A plugin crash degrades only its analyzer run.

The TypeScript control plane exposes local analysis sessions, paginated workspace views, SSE progress, health, readiness, and metrics. It binds to loopback by default. PostgreSQL is optional server-mode persistence; offline CLI analysis does not require it.

The JavaScript report viewer consumes Base64-encoded JSON embedded in a static HTML report. It uses DOM text nodes for untrusted values and makes no network requests. Sanitized exports are separate workspaces with their own manifest and hashes.

See [data flow](data-flow.md), [plugin system](plugin-system.md), and [workspace format](workspace-format.md) for component boundaries.
