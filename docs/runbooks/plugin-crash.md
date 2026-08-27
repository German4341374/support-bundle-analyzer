# Plugin crash

## Symptoms

One analyzer ends as `failed`, returns invalid JSONL, exceeds output limits, or exits unexpectedly while the session continues.

## Diagnosis

Check plugin version, protocol version, artifact type, bounded stderr, exit code, timeout, and memory pressure. Never include raw artifact content in an issue.

## Immediate mitigation

Disable the plugin for the session and continue with built-in analyzers. Treat its findings as unavailable, not negative evidence.

## Recovery

Reproduce with a synthetic fixture, fix the plugin, execute contract tests, and rerun into a new workspace.

## Prevention

Validate manifests, pin plugin versions, use direct process spawning, cap stdout/stderr/findings, and test crash/timeout/cancellation paths.
