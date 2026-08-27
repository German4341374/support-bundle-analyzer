# Benchmark results

## Streaming log analyzer baseline

Measured on 27 August 2026 from commit work-in-progress before publication.

| Field | Value |
|---|---|
| OS | Windows 10, amd64 |
| CPU | Intel Core i5-12400F, 6 cores / 12 logical processors |
| Memory | 31.8 GiB physical RAM |
| Go | 1.26.7 |
| Input | 1 MiB deterministic repeated application log |
| Command | `go test -run '^$' -bench '^BenchmarkAnalyzeLog1MiB$' -benchmem -count=3 ./internal/analyze` |

Raw measurements:

```text
BenchmarkAnalyzeLog1MiB-12    2    788467050 ns/op    45399608 B/op    448423 allocs/op
BenchmarkAnalyzeLog1MiB-12    2    786500900 ns/op    45405868 B/op    448404 allocs/op
BenchmarkAnalyzeLog1MiB-12    2    785330850 ns/op    45426160 B/op    448410 allocs/op
```

The median is 786.5 ms per MiB, approximately 1.27 MiB/s, with about 43.3 MiB allocated per operation. This is a baseline, not a universal throughput claim. The fixture intentionally triggers every line through timestamp parsing, privacy patterns, fingerprinting, evidence aggregation and timeline creation.

The allocation rate identifies a concrete optimization target: avoid repeated whole-line regular-expression copies and bound retained timeline payload earlier. Any future optimization must preserve grouping, privacy detection and archive safety tests, and this report must be regenerated on the same documented fixture.

Large-bundle peak RSS and PostgreSQL `EXPLAIN ANALYZE` results are not yet published because those workloads have not been measured in a controlled environment. Query templates exist under `database/queries` for the next measurement pass.
