# Tiered Node Fallback Scheduler

`tiered-node-scheduler` is a production-ready custom Kubernetes scheduler developed using the **Kubernetes Scheduler Framework** (v1.30.0).

The plugin implements a two-tier scheduling policy (*Graceful Tier Fallback*):
1. **Primary Node Pool:** Dedicated nodes identified by a configured key-value label selector (e.g., `pool: primary` or `tier: priority`).
2. **Secondary / Fallback Pool:** All remaining nodes within the cluster.

---

## Architecture & Scheduling State Flow

```text
       New Pod (spec.schedulerName: tiered-node-scheduler)
                                │
                                ▼
                     [ FilterPlugin Hook ]
                                │
                   Does Node have Primary Label?
                               / \
                         YES  /   \  NO (Fallback Node)
                             /     \
                            ▼       ▼
              [ Success: Schedule ]   Calculate Pod Age:
                                      time.Since(pod.CreationTimestamp)
                                            / \
               Age < fallbackTimeoutSeconds/   \  Age >= fallbackTimeoutSeconds
                                          /     \
                                         ▼       ▼
            [ Unschedulable: Grace Period ]     [ Success: Permit Placement ]
```

- When the primary pool is saturated, candidate pods enter the scheduler backoff queue with status `Unschedulable` for `fallbackTimeoutSeconds` (default: `20s`).
- Once the grace period expires, the filter permits scheduling onto secondary / fallback nodes.

---

## Local Development & Compilation on macOS (Apple Silicon / Intel)

### Prerequisites
- Go 1.22+
- Docker Desktop or OrbStack
- GNU Make

### Build Targets

```bash
# Code formatting and automated tests
make fmt
make test

# Build native binary for macOS (Darwin arm64 or amd64)
make build

# Cross-compile statically linked Linux binaries
make build-linux-amd64
make build-linux-arm64

# Local container build
make docker-build IMAGE_REPO=ghcr.io/gorizond/tiered-node-scheduler VERSION=dev
```

---

## Installation via Helm

### 1. Add Helm Repository from GitHub Pages

```bash
helm repo add tiered-scheduler https://gorizond.github.io/tiered-node-scheduler
helm repo update
```

### 2. Deploy the Scheduler Chart

```bash
helm upgrade --install tiered-node-scheduler tiered-scheduler/tiered-node-scheduler \
  --namespace kube-system \
  --set replicaCount=2 \
  --set pluginConfig.fallbackTimeoutSeconds=20 \
  --set pluginConfig.primaryNodeLabelKey="pool" \
  --set pluginConfig.primaryNodeLabelValue="primary"
```

---

## Workload Example

To route pods to this scheduler, set `spec.schedulerName: tiered-node-scheduler`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workload-fallback-demo
  namespace: default
  labels:
    app: demo
spec:
  replicas: 5
  selector:
    matchLabels:
      app: demo
  template:
    metadata:
      labels:
        app: demo
    spec:
      schedulerName: tiered-node-scheduler
      containers:
        - name: app
          image: cgr.dev/chainguard/static:latest
          resources:
            requests:
              cpu: 500m
              memory: 256Mi
```

### Verify Node Labeling & Events

```bash
# Label primary nodes
kubectl label node node-1 pool=primary
kubectl label node node-2 pool=primary

# Inspect scheduling events
kubectl get events --sort-by='.metadata.creationTimestamp' -A | grep tiered-node-scheduler
```
