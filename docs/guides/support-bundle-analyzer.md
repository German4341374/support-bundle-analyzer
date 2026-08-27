# Support Bundle Analyzer workflow

Support Bundle Analyzer is an offline workbench for ZIP/TAR diagnostic archives, application logs, HAR traces, Windows Event XML, JVM diagnostics, and PHP/web logs.

The useful unit is the investigation workspace: a source hash, artifact inventory, analyzer-run records, findings with evidence, a normalized timeline, privacy matches, and a static report. This makes a review repeatable between engineers.

```bash
support-bundle-analyzer generate-demo demo.zip
support-bundle-analyzer analyze demo.zip --output demo-investigation
support-bundle-analyzer report demo-investigation --output refreshed-report
```

Use `diff` when a known-healthy bundle exists. Differences provide direction but do not prove causality. Use `redact` only to create a separate sharing copy; never replace the evidence workspace.

The default Go pipeline currently wires built-in log and HAR rules. External analyzers can be executed through the documented JSONL protocol and are being integrated into automatic discovery after the engineering preview.
