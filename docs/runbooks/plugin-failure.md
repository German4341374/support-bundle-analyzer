# Runbook: plugin failure or timeout

1. Check `analyzer-runs.json` for status, duration and bounded warning text.
2. Confirm the plugin protocol version matches the core.
3. Run the plugin against a minimal synthetic artifact using the documented JSONL request.
4. Check CPU, memory and file limits; do not increase them without capacity review.
5. Disable the failing optional plugin and rerun. Built-in analyzer results remain valid.
6. If the plugin produced malformed or oversized output, treat it as untrusted and open a defect with the synthetic reproduction.
