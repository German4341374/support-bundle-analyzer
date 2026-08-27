INSERT INTO artifact_types (name, description) VALUES
    ('generic-log', 'Unstructured or semi-structured text log'),
    ('json-log', 'JSON or JSON Lines application log'),
    ('har', 'HTTP Archive document'),
    ('windows-event-xml', 'Exported Windows Event XML'),
    ('jvm-thread-dump', 'JVM thread dump'),
    ('jvm-gc-log', 'JVM garbage-collection log'),
    ('php-log', 'PHP or PHP-FPM log'),
    ('nginx-access', 'Nginx access log'),
    ('nginx-error', 'Nginx error log'),
    ('env-config', 'Environment-style configuration')
ON CONFLICT (name) DO NOTHING;

INSERT INTO analyzers (name, version, protocol_version, metadata) VALUES
    ('builtin-log-har', '0.1.0', '1', '{"runtime":"go"}'),
    ('log-intelligence-python', '0.1.0', '1', '{"runtime":"python"}'),
    ('windows-diagnostics-analyzer', '0.1.0', '1', '{"runtime":"dotnet"}'),
    ('jvm-diagnostics-analyzer', '0.1.0', '1', '{"runtime":"java"}'),
    ('php-web-diagnostics-analyzer', '0.1.0', '1', '{"runtime":"php"}')
ON CONFLICT (name) DO UPDATE SET version = EXCLUDED.version, metadata = EXCLUDED.metadata, updated_at = now();
