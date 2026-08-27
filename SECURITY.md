# Security policy

## Supported versions

Security fixes are applied to the latest released minor version. The repository is currently pre-1.0; interfaces may evolve with documented migration notes.

## Reporting a vulnerability

Use GitHub private vulnerability reporting for this repository. Do not open a public issue for a suspected vulnerability and do not attach real support bundles, secrets, exploit payloads containing third-party data, or customer information.

Include the affected version or commit, threat scenario, minimal synthetic reproduction, expected impact, and any proposed mitigation. You should receive an acknowledgement within five business days. No bounty or response-time guarantee is offered.

## Security boundaries

The archive parser, workspace manager, redaction pipeline, static-report renderer, API remote-mode guard and plugin process boundary are security-sensitive. Bundle content is untrusted data and must never become executable code. A finding is evidence, not a guarantee of root cause or safety.

See [`docs/security/threat-model.md`](docs/security/threat-model.md) for assumptions, abuse cases and residual risks.
