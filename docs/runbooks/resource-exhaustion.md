# Runbook: resource exhaustion

1. Stop accepting new analyses and allow the current safe point to complete.
2. Inspect configured maximum files, total bytes, single-file bytes, timeline events and findings.
3. Confirm free disk space before retaining a partial workspace.
4. Do not raise limits solely to accept an unknown archive. Reproduce with a smaller synthetic subset first.
5. Remove failed staging directories only after preserving error metadata required by your incident process.
6. Restart the service and verify `/health`, `/ready`, a small analysis and report generation.
