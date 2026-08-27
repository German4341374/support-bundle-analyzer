# Corrupt workspace

## Symptoms

Manifest parsing fails, referenced artifacts are missing, or stored hashes do not match.

## Diagnosis

Confirm schema version, filesystem errors, interrupted writes, disk capacity, and whether another process modified the workspace.

## Immediate mitigation

Mark the workspace read-only and stop using it for evidence or sharing.

## Recovery

Re-run analysis from the immutable source archive into a new output directory and compare source SHA-256 values.

## Prevention

Use atomic staging/finalization, private permissions, immutable completed workspaces, backups for required evidence, and disk monitoring.
