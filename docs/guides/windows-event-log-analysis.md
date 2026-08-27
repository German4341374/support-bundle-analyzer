# Windows Event Log analysis

Export relevant events as Windows Event XML and include them in the bundle. Binary EVTX is recognized but intentionally delegated to mature specialist tooling.

The Windows analyzer extracts Event ID, provider, channel, level, creation time, computer, user, and message. It groups evidence for authentication failures, unexpected reboots, service or application crashes, disk warnings, and network warnings.

Interpret an Event ID in context. A service crash event can confirm that a process stopped at a time, but it does not prove why. Correlate it with adjacent application logs, maintenance activity, disk state, and repeated stack traces.

Usernames and computer names may be personal or infrastructure-sensitive. Use strict redaction before sharing, and retain the original XML only in the private workspace.
