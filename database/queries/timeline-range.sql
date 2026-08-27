EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT event_timestamp, component, severity, message
FROM timeline_events
WHERE session_id = :'session_id'::uuid
  AND event_timestamp >= :'from_timestamp'::timestamptz
  AND event_timestamp < :'to_timestamp'::timestamptz
ORDER BY event_timestamp, id
LIMIT 500;
