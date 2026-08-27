# Plugin development

A plugin reads one JSON object per line from standard input and writes JSON objects per line to standard output. Protocol version `1` requests identify an artifact and include context such as the workspace root and timezone. Plugins return `finding`, `timeline`, `inventory`, `warning`, `summary` or `error` records defined by the versioned schemas.

Plugins must stream large artifacts, cap their own internal collections, avoid network access by default, never execute artifact content and never write matched secrets to logs. `plugin.json` declares name, version, protocol version, executable, supported artifact types, capabilities, timeout and maintainer.

The core starts a trusted executable with a fixed argv array, sends requests through stdin, validates JSONL output and enforces time/output/finding limits. A plugin crash degrades its analyzer run; it does not invalidate already committed results from other analyzers.
