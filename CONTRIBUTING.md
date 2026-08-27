# Contributing

Thank you for improving Support Bundle Analyzer. Contributions should preserve its privacy-first, evidence-based design.

## Before opening a change

1. Search existing issues and discussions.
2. For a new analyzer or protocol change, open a design issue first.
3. Use only synthetic fixtures. Never upload customer bundles, credentials, personal data, or production hostnames.
4. Read `AGENTS.md`, the architecture overview, and the threat model.

## Development

On Linux or WSL2:

```bash
./scripts/bootstrap.sh
./scripts/verify.sh
```

Optional runtimes are verified by their focused commands in the README. Add tests in the same runtime as the change. Archive parser changes require adversarial and fuzz coverage. Schema changes require a new versioned migration; do not edit an already released migration.

## Commits and pull requests

Use Conventional Commits, for example `feat(analyzer): detect DNS failures` or `fix(archive): reject normalized duplicate paths`. Keep commits reviewable and explain the security impact. A pull request must describe behavior, tests run, risks, documentation changes, and whether schemas or migrations changed.

All required CI checks must pass. Maintainers may ask for a smaller fixture or narrower dependency before accepting a change.
