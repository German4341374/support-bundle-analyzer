# PostgreSQL persistent mode

The database stores analysis metadata and normalized results for the optional server mode. Raw bundle bytes remain outside PostgreSQL and should be placed on encrypted storage with a retention policy.

Migrations are versioned and run in transactions. Apply `001_initial.up.sql` to an empty database. Use `001_initial.down.sql` only during development; production rollback should normally restore a verified backup or deploy a forward corrective migration.

Transaction boundaries are intentionally short:

1. bundle and analysis-session creation;
2. artifact manifest insertion;
3. each analyzer result batch;
4. final session state transition.

The whole analysis is not one transaction because plugin execution can take minutes. The session state records partial progress, while each committed batch remains internally consistent.

The query files contain `EXPLAIN (ANALYZE, BUFFERS)` templates. They require a populated database and explicit `psql` variables. No performance output is committed until it has been measured against a documented dataset and environment.
