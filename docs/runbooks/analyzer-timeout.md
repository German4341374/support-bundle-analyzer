# Analyzer timeout

## Symptoms

An analyzer run ends as `timed_out` while the overall session completes with warnings.

## Diagnosis

Check artifact size, configured timeout, plugin stderr summary, CPU/memory pressure, and whether the parser is streaming.

## Immediate mitigation

Keep the artifact quarantined, reduce the diagnostic time window, or skip the affected analyzer. Never remove all resource limits.

## Recovery

Reproduce with a synthetic subset, fix unbounded parsing, then rerun into a new workspace.

## Prevention

Benchmark representative sizes, enforce cancellation, cap output, and test deliberate timeout behavior.
