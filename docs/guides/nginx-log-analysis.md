# nginx log analysis

Include a bounded incident window from both access and error logs. Preserve the configured timezone and log format alongside the sample.

Look for changes in status distribution, repeated upstream timeouts, connection refusals, permission failures, large response times, and correlation identifiers. A sequence of `502` responses suggests the proxy could not obtain a valid upstream response; it does not distinguish an unavailable upstream from protocol, timeout, or configuration failures without error-log evidence.

```bash
support-bundle-analyzer analyze nginx-support.zip --output nginx-investigation
```

Avoid publishing client addresses, authenticated paths, query strings, cookies, referrers, or user agents without review. Custom nginx formats may require a new parser fixture; unmatched lines remain available as artifacts instead of being silently discarded.
