# Archive limit triggered

## Symptoms

Analysis rejects a bundle for file count, entry size, total expanded size, nesting, path length, or compression ratio.

## Diagnosis

Record the rule and effective limit from the safe error and manifest warning. Do not extract the archive manually to inspect it.

## Immediate mitigation

Quarantine the bundle and request a smaller bundle or an incident-window subset from the sender.

## Recovery

Validate a regenerated archive with the same limits. Raise a limit only after estimating storage and memory impact in an isolated environment.

## Prevention

Document collection limits, reject nested bulk exports, and retain adversarial archive tests in CI.
