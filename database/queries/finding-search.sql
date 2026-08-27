EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, title, ts_rank_cd(search_vector, websearch_to_tsquery('english', :'query')) AS rank
FROM findings
WHERE session_id = :'session_id'::uuid
  AND search_vector @@ websearch_to_tsquery('english', :'query')
ORDER BY rank DESC, id
LIMIT 50;
