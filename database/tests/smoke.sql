\set ON_ERROR_STOP on

BEGIN;
\ir ../migrations/001_initial.up.sql
\ir ../seed/development.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM artifact_types) < 10 THEN
        RAISE EXCEPTION 'artifact type seed is incomplete';
    END IF;
    IF (SELECT count(*) FROM analyzers) < 5 THEN
        RAISE EXCEPTION 'analyzer seed is incomplete';
    END IF;
END $$;

ROLLBACK;
