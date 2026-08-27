# How to analyze a support bundle safely

Start with a copy of the archive and calculate its hash before extraction. The analyzer performs this automatically and records the SHA-256 in `manifest.json`.

```bash
support-bundle-analyzer analyze incident.zip --output investigation --timezone UTC
```

Review the output in this order: manifest warnings, failed analyzer runs, high-severity findings, timeline clusters, then artifacts around the evidence ranges. A finding means an observation matched a deterministic rule; it is not automatically the incident root cause.

If the archive triggers a path, link, size, or compression-ratio limit, stop and follow the unsafe-archive runbook. Do not bypass limits by manually extracting the bundle on a workstation. If source timestamps omit offsets, pass the source timezone and treat cross-host ordering as approximate until clock synchronization is verified.

Before sharing any output, create a strict redacted copy and inspect its redaction manifest. Keep the original workspace private and immutable.
