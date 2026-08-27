# Runbook: unsafe or malformed archive

1. Stop analysis and preserve the original archive hash and error code.
2. Do not open the bundle with a general-purpose archive manager on a workstation.
3. Quarantine it according to your organization's incident procedure.
4. Ask the sender to regenerate the bundle using a documented export process.
5. If investigation is necessary, reproduce with a synthetic archive and report privately under `SECURITY.md`.

Do not bypass path, link, size, count or compression-ratio controls to make a customer bundle pass.
