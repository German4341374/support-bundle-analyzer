# How to redact diagnostic logs

Keep the original investigation private and produce a separate strict export:

```bash
support-bundle-analyzer redact investigation --profile strict --output sanitized-investigation
```

Review detected categories and replacement counts. Search for organization-specific identifiers that generic rules cannot know: tenant IDs, case numbers, internal hostnames, proprietary tokens, and customer-defined fields.

Stable pseudonyms preserve relationships such as the same IP appearing across two services. They do not make the report anonymous under every privacy regime. Very short secrets, encrypted blobs, binary files, and unusual encodings may evade detection.

Share only the minimum artifacts needed for the question, use an approved transfer channel, set a retention period, and record who received the export.
