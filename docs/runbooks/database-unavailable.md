# Database unavailable

## Symptoms

Readiness fails, persistent API operations return dependency errors, or PostgreSQL connections time out. Offline CLI analysis remains available.

## Diagnosis

Check PostgreSQL health, connection limits, DNS, TLS policy, credentials, storage, and recent migrations. Do not print `DATABASE_URL`.

## Immediate mitigation

Stop accepting persistent jobs and direct users to offline CLI mode. Do not retry writes without an idempotency boundary.

## Recovery

Restore PostgreSQL, run migration smoke checks on a disposable connection, restart the API, and confirm `/ready` plus a read/write synthetic session.

## Prevention

Alert on pool exhaustion and migration failure, back up the database, test restore, and keep retention cleanup bounded.
