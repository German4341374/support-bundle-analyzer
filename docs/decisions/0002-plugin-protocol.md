# ADR 0002: Use JSON Lines subprocess plugins

Status: Accepted

## Options considered

- gRPC provides typed streaming but requires generated clients and a network listener.
- HTTP is easy to inspect but introduces ports, lifecycle management, and a larger attack surface.
- JSON Lines over subprocess standard streams is portable, inspectable, and works offline.

## Decision

The core spawns a configured executable with an argument array, never a concatenated shell command. One request is written as JSON and responses are read one JSON object per line. The runner enforces cancellation, timeout, stdout/stderr limits, a finding limit, and schema validation. A plugin failure degrades only that analyzer run.

Protocol versions are independent from product versions. Version `1` is described by the schemas in `packages/schemas/v1`.

