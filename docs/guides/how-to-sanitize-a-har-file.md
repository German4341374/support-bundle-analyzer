# How to sanitize a HAR file

Analyze first so the original source hash and privacy findings are preserved in a private workspace. Then create a separate export:

```bash
support-bundle-analyzer redact har-investigation --profile strict --output har-sanitized
```

Inspect `har-sanitized/redaction-manifest.json` and search the exported files for every synthetic token you expect to be removed. Verify Authorization, Cookie, Set-Cookie, URL query values, request bodies, emails, IP addresses, and user paths.

Stable placeholders such as `EMAIL_001` help retain relationships inside one report. They are not encryption and must not be used as reversible identifiers. Binary bodies are excluded when the tool cannot safely rewrite them.

Do not delete the private source until the investigation and retention requirements are complete. Do not assume automated sanitization finds every organization-specific identifier.
