EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, severity, rule_id, title, occurrences
FROM findings
WHERE session_id = :'session_id'::uuid
ORDER BY CASE severity
    WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5
END, id
LIMIT 100;
