#!/usr/bin/env bash
set -euo pipefail

cluster="sba-smoke"
cleanup() { kind delete cluster --name "${cluster}" >/dev/null 2>&1 || true; }
trap cleanup EXIT

kind create cluster --name "${cluster}" --wait 120s
docker build -t support-bundle-analyzer:smoke .
kind load docker-image support-bundle-analyzer:smoke --name "${cluster}"
kubectl create namespace sba
kubectl -n sba create secret generic support-bundle-analyzer-auth --from-literal=access-token=local-smoke-token-1234567890
helm upgrade --install sba deploy/helm/support-bundle-analyzer --namespace sba \
  --set image.repository=support-bundle-analyzer \
  --set image.tag=smoke \
  --set image.pullPolicy=Never \
  --wait --timeout 180s
kubectl -n sba rollout status deployment/support-bundle-analyzer --timeout=120s
kubectl -n sba port-forward service/support-bundle-analyzer 18080:8080 >.tmp/kind-port-forward.log 2>&1 &
forward_pid=$!
trap 'kill ${forward_pid} >/dev/null 2>&1 || true; cleanup' EXIT
for _ in {1..30}; do
  if curl --fail --silent http://127.0.0.1:18080/health >/dev/null; then
    echo "kind smoke test passed"
    exit 0
  fi
  sleep 1
done
echo "kind smoke test failed" >&2
exit 1
