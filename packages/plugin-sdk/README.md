# Analyzer plugin SDK

Protocol version 1 is transport-neutral JSON Lines over stdin/stdout. `packages/schemas/v1/plugin.schema.json` validates the manifest, while finding and timeline schemas validate normalized records.

Minimal request:

```json
{"protocolVersion":"1","analysisId":"00000000-0000-4000-8000-000000000001","artifact":{"path":"logs/api.log","type":"generic-log","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"context":{"timezone":"UTC","workspaceRoot":"/private/workspace/artifacts"}}
```

Minimal response:

```json
{"type":"finding","finding":{"ruleId":"EXAMPLE_FAILURE","severity":"medium","title":"Example evidence","summary":"A synthetic failure marker was observed.","confidence":"strong","evidence":[{"artifact":"logs/api.log","lineStart":12,"lineEnd":12}]}}
```

Do not place diagnostic text on stdout unless it is valid protocol JSON. Use stderr for bounded operational diagnostics and never include raw secrets or complete log lines there.
