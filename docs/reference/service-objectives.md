# Example service objectives

These are example deployment policies for self-hosted server mode. They are not claims about achieved production reliability.

- 99.9% successful API requests over 30 days, excluding valid client errors.
- 99% of accepted small-bundle jobs begin processing within 30 seconds.
- 99% of health checks respond within 500 ms while dependencies are healthy.
- Zero high-cardinality artifact paths or detected secret values in metrics.

Suggested indicators include API response status, analysis duration, analyzer failures, queue depth, report generation duration, and retention cleanup failures. Operators should define bundle-size classes because processing latency depends heavily on archive composition and storage speed.

An alert should link to the matching runbook and preserve privacy: identifiers may be logged, but filenames, raw evidence, tokens, and customer payloads must not become metric labels or alert text.
