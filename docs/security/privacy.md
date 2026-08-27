# Privacy model

Support bundles may contain passwords, bearer tokens, cookies, private keys, connection strings, email addresses, IP addresses, usernames, hostnames, and application payloads.

Local mode reads the selected archive and writes only to the selected workspace. Telemetry, analytics, cloud upload, remote processing, and automatic network access are disabled. The project does not claim that secret or PII detection is complete.

Standard redaction targets credentials and high-confidence secrets. Strict redaction additionally pseudonymizes emails, IP addresses, phone-like values, and home-directory usernames. Stable placeholders preserve correlations inside one export but are not stable across independently created exports.

Binary artifacts are excluded from sanitized exports because safe rewriting cannot be guaranteed. Raw request bodies, HAR cookies, environment files, and configuration dumps require manual review even after automated redaction.

Local data remains until the user deletes it. Server operators must configure retention, restrict workspace permissions, encrypt storage where appropriate, and audit deletion. Never attach a real bundle to a public issue. Use GitHub private vulnerability reporting for security defects.
