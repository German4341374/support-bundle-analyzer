# Error taxonomy

| Code | Meaning | Operator action |
|---|---|---|
| `ARCHIVE_PATH_TRAVERSAL` | an entry resolves outside the extraction root | quarantine and regenerate the bundle |
| `ARCHIVE_UNSAFE_ENTRY` | a link, special file or duplicate normalized path was found | do not bypass the rejection |
| `ARCHIVE_LIMIT_EXCEEDED` | an entry or archive exceeds configured resource limits | verify capacity and request a smaller export |
| `ARCHIVE_UNSUPPORTED_FORMAT` | input magic/extension is not supported | convert with a trusted tool or use a supported export |
| `ARTIFACT_MALFORMED` | a structured artifact could not be parsed | inspect the source/export process |
| `ANALYZER_FAILED` | analyzer exited or returned invalid output | follow the plugin failure runbook |
| `ANALYZER_TIMEOUT` | analyzer exceeded its deadline | inspect artifact size and analyzer performance |
| `REPORT_GENERATION_FAILED` | normalized data could not be serialized or written | verify disk space and workspace permissions |
| `DATABASE_UNAVAILABLE` | optional persistent mode cannot reach PostgreSQL | use local CLI or follow database recovery |

API errors use `{ "error": { "code": "...", "message": "..." } }`. Internal stack traces and artifact contents are not returned to clients.
