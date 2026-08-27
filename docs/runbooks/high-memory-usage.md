# High memory usage

## Symptoms

The process approaches its container limit, slows under paging, or is terminated for out-of-memory use.

## Diagnosis

Record bundle size, file count, largest entry, analyzer, concurrency, and phase. Check whether an analyzer loaded an entire file or retained unbounded findings.

## Immediate mitigation

Cancel the job, reduce concurrency, lower input limits, and analyze a smaller time window. Preserve the source hash and failure metadata.

## Recovery

Reproduce with synthetic scale fixtures, replace whole-file processing with bounded streaming, and confirm memory with the benchmark suite.

## Prevention

Enforce archive, response, plugin-output, and finding limits; maintain container limits and regression benchmarks.
