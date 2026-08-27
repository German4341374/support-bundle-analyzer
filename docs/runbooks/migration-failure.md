# Migration failure

## Symptoms

Deployment stops during schema migration or the API reports an unsupported schema version.

## Diagnosis

Read the first PostgreSQL error, verify the applied migration set, available extensions, locks, disk space, and backup state.

## Immediate mitigation

Keep the new application version stopped. Do not edit the production schema manually or rerun a partially applied non-transactional migration.

## Recovery

Restore from a tested backup when required, reproduce on a disposable database, correct the forward migration, and validate smoke queries before redeployment.

## Prevention

Keep migrations transactional where PostgreSQL permits, test upgrade and rollback paths, use advisory deployment locks, and require backups before destructive changes.
