BEGIN;

CREATE TABLE artifact_types (
    id smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE CHECK (name ~ '^[a-z0-9][a-z0-9-]{1,63}$'),
    description text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE analyzers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    version text NOT NULL,
    protocol_version text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE bundles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sha256 char(64) NOT NULL UNIQUE CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    original_name text NOT NULL CHECK (length(original_name) BETWEEN 1 AND 1024),
    size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
    stored_path text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE analysis_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id uuid NOT NULL REFERENCES bundles(id) ON DELETE RESTRICT,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'completed_with_warnings', 'failed', 'cancelled')),
    tool_version text NOT NULL,
    schema_version text NOT NULL,
    timezone text NOT NULL DEFAULT 'UTC',
    settings jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(settings) = 'object'),
    warning_count integer NOT NULL DEFAULT 0 CHECK (warning_count >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    CONSTRAINT analysis_time_order CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at)
);

CREATE TABLE artifacts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    artifact_type_id smallint NOT NULL REFERENCES artifact_types(id) ON DELETE RESTRICT,
    relative_path text NOT NULL CHECK (relative_path !~ '(^/|(^|/)\.\.(/|$))'),
    sha256 char(64) NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, relative_path),
    UNIQUE (session_id, sha256, relative_path)
);

CREATE TABLE analyzer_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    analyzer_id uuid NOT NULL REFERENCES analyzers(id) ON DELETE RESTRICT,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'completed_with_warnings', 'failed', 'timed_out', 'skipped')),
    duration_ms bigint CHECK (duration_ms >= 0),
    warnings jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(warnings) = 'array'),
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE findings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    analyzer_run_id uuid REFERENCES analyzer_runs(id) ON DELETE SET NULL,
    rule_id text NOT NULL,
    severity text NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    category text NOT NULL,
    component text,
    title text NOT NULL,
    summary text NOT NULL,
    explanation text,
    confidence text NOT NULL CHECK (confidence IN ('strong', 'moderate', 'weak')),
    occurrences integer NOT NULL DEFAULT 1 CHECK (occurrences > 0),
    next_steps jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(next_steps) = 'array'),
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(summary, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(component, '') || ' ' || coalesce(rule_id, '')), 'C')
    ) STORED,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, rule_id, component, title)
);

CREATE TABLE evidence (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    finding_id uuid NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    artifact_id uuid NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    line_start integer CHECK (line_start IS NULL OR line_start > 0),
    line_end integer CHECK (line_end IS NULL OR line_end >= line_start),
    json_pointer text,
    event_timestamp timestamptz,
    excerpt text CHECK (excerpt IS NULL OR length(excerpt) <= 4096),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE timeline_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    artifact_id uuid REFERENCES artifacts(id) ON DELETE SET NULL,
    event_timestamp timestamptz NOT NULL,
    source text NOT NULL,
    component text,
    severity text NOT NULL CHECK (severity IN ('critical', 'error', 'warning', 'info', 'debug')),
    category text NOT NULL,
    message text NOT NULL,
    correlation_id text,
    evidence jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(evidence) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE redaction_findings (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    artifact_id uuid REFERENCES artifacts(id) ON DELETE CASCADE,
    kind text NOT NULL,
    match_count integer NOT NULL CHECK (match_count > 0),
    profile text NOT NULL CHECK (profile IN ('review-only', 'standard', 'strict')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, artifact_id, kind, profile)
);

CREATE TABLE reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    format text NOT NULL CHECK (format IN ('static-html', 'json', 'markdown', 'sarif')),
    sanitized boolean NOT NULL DEFAULT false,
    sha256 char(64) NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    relative_path text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, format, sanitized)
);

CREATE TABLE bundle_comparisons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    baseline_session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    incident_session_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    result jsonb NOT NULL CHECK (jsonb_typeof(result) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (baseline_session_id <> incident_session_id)
);

CREATE TABLE plugin_registry (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    version text NOT NULL,
    protocol_version text NOT NULL,
    executable text NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    manifest jsonb NOT NULL CHECK (jsonb_typeof(manifest) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (name, version)
);

CREATE TABLE audit_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_timestamp timestamptz NOT NULL DEFAULT now(),
    actor text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    request_id text,
    details jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(details) = 'object')
);

CREATE INDEX idx_sessions_status_created ON analysis_sessions (status, created_at DESC);
CREATE INDEX idx_artifacts_session_type ON artifacts (session_id, artifact_type_id);
CREATE INDEX idx_artifacts_type ON artifacts (artifact_type_id);
CREATE INDEX idx_runs_session_analyzer ON analyzer_runs (session_id, analyzer_id);
CREATE INDEX idx_findings_session_severity ON findings (session_id, severity);
CREATE INDEX idx_findings_analyzer ON findings (analyzer_run_id) WHERE analyzer_run_id IS NOT NULL;
CREATE INDEX idx_findings_search ON findings USING gin (search_vector);
CREATE INDEX idx_evidence_finding ON evidence (finding_id);
CREATE INDEX idx_timeline_session_time ON timeline_events (session_id, event_timestamp, id);
CREATE INDEX idx_timeline_correlation ON timeline_events (session_id, correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_redaction_session ON redaction_findings (session_id);
CREATE INDEX idx_audit_resource ON audit_events (resource_type, resource_id, event_timestamp DESC);

COMMIT;
