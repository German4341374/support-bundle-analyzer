# Kubernetes support bundle analysis

Collect only the namespaces and time window needed for the incident. Prefer manifests, events, pod descriptions, and bounded logs; do not include live service-account tokens or Secret values.

The current engineering preview classifies Kubernetes YAML, events, and logs and applies generic configuration/log analysis. Platform-specific correlation remains limited, so confirm observations with:

```bash
kubectl get pods -n <namespace>
kubectl describe pod -n <namespace> <pod>
kubectl get events -n <namespace> --sort-by=.lastTimestamp
```

Useful evidence includes Pending reasons, restart counts, failed probes, image pull failures, scheduling constraints, resource pressure, and rollout timestamps. A CrashLoopBackOff is a retry state, not a root cause; inspect the previous container logs and termination reason.

Sanitize cluster names, internal addresses, registry credentials, environment values, and annotations before sharing.
