# PHP-FPM log analysis

Collect PHP error logs, PHP-FPM logs, a sanitized `php.ini`, composer manifests, and reverse-proxy errors from the same incident window.

The PHP analyzer detects fatal errors, memory-limit exhaustion, maximum execution time, worker saturation, upstream timeouts, permission failures, repeated HTTP 5xx, and version metadata. It reports only what appears in the provided artifacts.

Correlate `server reached pm.max_children` with request latency and pool configuration. Treat `Allowed memory size exhausted` as evidence of memory exhaustion in a request, then inspect the named code path and workload rather than simply raising the limit.

Remove database URLs, Composer repository credentials, environment variables, session identifiers, document-root usernames, and customer paths before sharing.
