# Runbook: PostgreSQL or migration failure

1. Stop new server-mode analyses; CLI local-first operation does not require PostgreSQL.
2. Capture PostgreSQL health, migration version and application error code without logging credentials.
3. If a migration failed inside its transaction, fix the cause and rerun; PostgreSQL rolls back the incomplete transaction.
4. If data changed outside a migration transaction, restore the last verified backup to a new database and validate it before switching traffic.
5. Prefer a forward corrective migration. Use `001_initial.down.sql` only for disposable development databases.
6. Reconcile sessions left in `running` state and mark them failed or resume from a validated checkpoint once persistence is implemented.
