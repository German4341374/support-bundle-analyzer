# Design notes

1. **Why is Go the core runtime?** It provides a small deployable binary, strong streaming and concurrency primitives, and predictable resource control around untrusted archives.
2. **Why use plugins?** Platform formats need different ecosystems. A JSONL boundary keeps the security-sensitive core independent and makes crashes containable.
3. **Why JSON Lines instead of HTTP for plugins?** Local subprocess pipes avoid ports and service discovery, stream naturally and preserve language neutrality.
4. **What prevents Zip Slip?** Normalization rejects absolute, drive-qualified and traversal paths, then verifies the final relative path remains under the private root.
5. **How are archive bombs limited?** File count, per-file bytes, total bytes, compression ratio and output-copy limits are enforced before or during extraction.
6. **Why reject symlinks and hardlinks?** Their targets can escape the extraction root or alias unexpected host files.
7. **What makes a finding trustworthy?** It includes a stable rule, confidence, bounded summary and concrete artifact/line/JSON-pointer evidence.
8. **Why avoid claiming root cause?** A support bundle is partial evidence. A matching symptom can have several causes.
9. **How is XSS controlled?** Report data is Base64-encoded JSON and displayed through DOM `textContent`; bundle HTML/JavaScript is never evaluated.
10. **What is the redaction guarantee?** Risk reduction only. Known patterns are pseudonymized, binaries excluded, a manifest records actions, and human review remains mandatory.
11. **Why stable pseudonyms?** They preserve correlations such as one address appearing across files without exposing its original value.
12. **How does remote API mode differ?** Loopback is the default. Non-loopback binding requires explicit opt-in and bearer authentication.
13. **Why cursor pagination?** Findings and timeline data can be large; bounded stable pages avoid memory spikes and huge API responses.
14. **What happens when a plugin crashes?** The runner records failure or timeout, caps output and lets other analyzers and the session continue.
15. **Why short database transactions?** Plugin execution is slow. Committing manifests and analyzer batches separately preserves consistency without long locks.
16. **Why PostgreSQL full-text search?** It provides ranked, indexed search for persistent multi-session mode without an extra search service.
17. **What is local-first?** Core analysis, reporting and redaction work without a database, cloud account or external API.
18. **How is determinism improved?** Artifacts and normalized results are sorted, schemas/version metadata are recorded, and synthetic fixtures are deterministic.
19. **Why not parse heap dumps directly?** They are large specialized binary formats. Recognition plus guidance is safer and more honest than a shallow parser.
20. **How do you test security properties?** Table tests, adversarial archive fixtures, fuzz targets, XSS payload tests and clean end-to-end analysis.
21. **Why are binaries excluded from sanitized exports?** Arbitrary rewriting corrupts binaries and regex scanning cannot establish that they are safe to share.
22. **What happens if PostgreSQL is unavailable?** Server persistence degrades, but the standalone CLI remains usable; recovery follows a documented runbook.
23. **How is supply-chain risk addressed?** Minimal runtime dependencies, lockfiles, pinned workflow actions, dependency review, CodeQL, secret scanning, SBOM and image scanning.
24. **What would you build next?** Automatic plugin discovery/orchestration and durable API sessions because they unlock the polyglot architecture without weakening core boundaries.
