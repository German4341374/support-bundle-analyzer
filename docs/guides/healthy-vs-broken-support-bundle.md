# Compare healthy and broken support bundles

The repository includes deterministic healthy and database-outage scenarios so the complete comparison can be reproduced without private data:

```bash
support-bundle-analyzer generate-demo healthy.zip --scenario healthy
support-bundle-analyzer generate-demo outage.zip --scenario database-outage
support-bundle-analyzer analyze healthy.zip --output healthy-workspace --timezone UTC
support-bundle-analyzer analyze outage.zip --output outage-workspace --timezone UTC
support-bundle-analyzer diff healthy-workspace outage-workspace --output comparison.json
```

The current synthetic comparison reports four new outage findings, three changed artifacts, and a timeline increase from one to four events. Those values were measured from the bundled scenarios on 28 August 2026; rerun the commands after changing rules or fixtures.

For a real investigation, capture the same artifact set, configuration scope, timezone, and approximate workload for both environments. Analyze each bundle independently, then compare the completed workspaces:

```bash
support-bundle-analyzer diff healthy-workspace broken-workspace --output comparison.json
```

Review added or removed artifact types, changed findings, severity counts, component inventory, configuration fingerprints, and timeline density. A difference is a lead, not proof. Version drift may be intentional, and identical bundles can still hide an external dependency failure.

Do not compare raw customer data across unrelated tenants. Apply equivalent redaction profiles if the comparison leaves the private investigation boundary, and preserve both source hashes so another engineer can reproduce the result.
