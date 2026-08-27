# How to analyze HAR files

HAR files contain request URLs, headers, cookies, timings, and often request or response bodies. Treat them as credentials until reviewed.

Place the HAR inside a synthetic or private bundle and run:

```bash
support-bundle-analyzer analyze browser-diagnostics.zip --output har-investigation
```

Use the report to examine request count, domains, HTTP failures, authentication responses, slow requests, redirect evidence, content types, transfer sizes, and available timing fields. A `401` shows that authentication was rejected; it does not by itself explain whether the token, audience, clock, proxy, or server policy was responsible.

Look for clusters around the incident time and compare them with application logs. Missing HAR timing fields remain unknown rather than being inferred. Embedded HTML or JavaScript is displayed as text and is never executed.

Always perform a strict privacy review before exporting a HAR-derived report. Authorization, Cookie, Set-Cookie, tokens, query values, forms, and bodies are especially sensitive.
