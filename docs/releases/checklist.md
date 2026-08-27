# Release checklist

A release is allowed only after every applicable item is backed by a successful command or GitHub check. An unchecked item blocks the release.

## Quality gates

- [ ] `make verify` passes from a clean checkout on Linux or WSL2.
- [ ] Unit, contract, integration, adversarial archive, and browser E2E tests pass.
- [ ] PostgreSQL migrations apply to a new database and the database smoke test passes.
- [ ] Docker Compose reaches healthy state.
- [ ] The Helm chart passes the kind smoke test.
- [ ] Terraform formatting, initialization, and validation pass.
- [ ] Gitleaks, CodeQL, Trivy, dependency audit, and container scanning pass.
- [ ] No unresolved high or critical security advisory remains.

## Release assets

- [ ] Version and changelog follow Semantic Versioning and Keep a Changelog.
- [ ] CLI binaries are produced for the documented platforms.
- [ ] SHA-256 checksums are generated and verified.
- [ ] SPDX SBOM and vulnerability-scan artifacts are attached.
- [ ] Documentation commands match the packaged artifacts.
- [ ] Screenshots and the synthetic Pages demo reflect the release candidate.
- [ ] Known limitations and upgrade notes are explicit.

## Publication

- [ ] The release commit is signed or protected by the repository ruleset.
- [ ] The release workflow succeeds for the version tag.
- [ ] The published archive and container digest match CI output.
- [ ] A post-release smoke test succeeds using only public artifacts.

Version `1.0.0` must not be published while the gaps in [project status](../project-status.md) remain open.
