# AxonHub Operator — Usage Guide

This document describes how to install, configure, and use the AxonHub
Operator in a Kubernetes cluster.

> **Status:** The `AxonHub` CRD and API types are defined. Controller
> reconciliation logic is under active development.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Option 1: Helm Chart (OCI / GHCR)](#option-1-helm-chart-oci--ghcr)
  - [Option 2: Helm Chart (Local)](#option-2-helm-chart-local)
  - [Option 3: Kustomize (Make targets)](#option-3-kustomize-make-targets)
  - [Option 4: Single YAML Bundle](#option-4-single-yaml-bundle)
- [Configuration](#configuration)
- [Verifying the Installation](#verifying-the-installation)
- [Deploying AxonHub Instances](#deploying-axonhub-instances)
  - [Embedded Postgres (default)]#embedded-postgres-default)
  - [External Postgres (BYO database)](#external-postgres-byo-database)
  - [Multiple Instances](#multiple-instances)
- [AxonHub CRD Reference](#axonhub-crd-reference)
- [Metrics & Observability](#metrics--observability)
- [Upgrading](#upgrading)
- [Uninstalling](#uninstalling)
- [Development](#development)

---

## Prerequisites

- Kubernetes cluster **v1.27+**
- `kubectl` configured to access your cluster
- Helm **v3.12+** (only required for Helm-based installs)
- (Optional) Prometheus Operator installed if you enable the ServiceMonitor

---

## Installation

### Option 1: Helm Chart (OCI / GHCR)

The chart is published as an OCI artifact to GitHub Container Registry.

**Latest (from `main` branch):**

```sh
helm install axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 0.0.0-latest \
  --namespace axon-operator-system \
  --create-namespace
```

**Tagged release (e.g. `v1.2.3`):**

```sh
helm install axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 1.2.3 \
  --namespace axon-operator-system \
  --create-namespace
```

**With a values file:**

```sh
helm install axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 1.2.3 \
  --namespace axon-operator-system \
  --create-namespace \
  -f my-values.yaml
```

### Option 2: Helm Chart (Local)

Useful for development and testing against a local checkout.

```sh
helm install axon-operator ./charts/axon-operator \
  --namespace axon-operator-system \
  --create-namespace
```

### Option 3: Kustomize (Make targets)

Build and push the image, then deploy with kustomize:

```sh
make docker-build docker-push IMG=<your-registry>/axon-operator:tag
make install                       # Install CRDs
make deploy IMG=<your-registry>/axon-operator:tag
```

### Option 4: Single YAML Bundle

Download the consolidated `install.yaml` from a GitHub release and apply it:

```sh
kubectl apply -f https://raw.githubusercontent.com/amghazanfari/axon-operator/<tag>/dist/install.yaml
```

---

## Configuration

The full list of configurable values is in
[`charts/axon-operator/values.yaml`](../charts/axon-operator/values.yaml).
The most common options are listed below.

| Parameter                           | Description                                          | Default                              |
|-------------------------------------|------------------------------------------------------|--------------------------------------|
| `replicaCount`                      | Number of controller replicas                        | `1`                                  |
| `image.repository`                  | Controller image repository                          | `ghcr.io/amghazanfari/axon-operator` |
| `image.tag`                         | Image tag (defaults to chart `appVersion`)           | `""`                                 |
| `image.pullPolicy`                  | Image pull policy                                    | `IfNotPresent`                       |
| `namespace`                         | Namespace for the operator                           | `axon-operator-system`               |
| `leaderElection.enabled`            | Enable leader election                               | `true`                               |
| `metrics.enabled`                   | Expose the metrics endpoint                          | `true`                               |
| `metrics.secure`                    | Serve metrics over HTTPS                             | `true`                               |
| `metrics.service.port`              | Metrics service port                                 | `8443`                               |
| `networkPolicy.enabled`             | Restrict metrics ingress to `metrics: enabled` ns    | `false`                              |
| `prometheus.monitor.enabled`        | Create a Prometheus ServiceMonitor                   | `false`                              |
| `resources.limits.cpu`              | CPU limit                                            | `500m`                               |
| `resources.limits.memory`           | Memory limit                                         | `128Mi`                              |
| `resources.requests.cpu`            | CPU request                                          | `10m`                                |
| `resources.requests.memory`         | Memory request                                       | `64Mi`                               |
| `nodeSelector`                      | Node labels for scheduling                           | `{}`                                 |
| `tolerations`                       | Tolerations for scheduling                           | `[]`                                 |
| `affinity`                          | Affinity rules for scheduling                        | `{}`                                 |
| `extraArgs`                         | Extra args appended to the manager binary            | `[]`                                 |
| `extraEnv`                          | Extra environment variables                          | `[]`                                 |
| `extraVolumes`                      | Extra volumes for the pod                            | `[]`                                 |
| `extraVolumeMounts`                 | Extra volume mounts for the container                | `[]`                                 |
| `podAnnotations`                    | Extra pod annotations                                | `{}`                                 |
| `podLabels`                         | Extra pod labels                                     | `{}`                                 |
| `serviceAccount.create`             | Create a dedicated service account                   | `true`                               |
| `serviceAccount.annotations`        | Annotations for the service account                  | `{}`                                 |
| `serviceAccount.name`               | Override the service account name                    | `""`                                 |

> **Note:** The operator Helm chart does **not** deploy a Postgres database.
> Each `AxonHub` CR manages its own database lifecycle (see below).

### Example: custom values

```yaml
# my-values.yaml
replicaCount: 2
image:
  tag: "1.2.3"
resources:
  limits:
    cpu: 1000m
    memory: 256Mi
nodeSelector:
  node-role.kubernetes.io/control-plane: ""
prometheus:
  monitor:
    enabled: true
    additionalLabels:
      release: prometheus
```

```sh
helm upgrade axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 1.2.3 \
  --namespace axon-operator-system \
  -f my-values.yaml
```

---

## Verifying the Installation

Check that the controller pod is running:

```sh
kubectl get pods -n axon-operator-system
```

Expected output:

```
NAME                                              READY   STATUS    RESTARTS   AGE
axon-operator-controller-manager-xxxxx            1/1     Running   0          30s
```

Check the operator logs:

```sh
kubectl logs -n axon-operator-system -l control-plane=controller-manager -c manager -f
```

Verify the CRD is registered:

```sh
kubectl get crd axonhubs.axonhub.looplj.com
```

---

## Deploying AxonHub Instances

The `AxonHub` CR is namespaced — you can create multiple isolated instances
in different namespaces. Each instance gets its own dedicated Postgres
database (when using the embedded mode) or connects to an external one.

### Embedded Postgres (default)

By default, the operator creates a single-node Postgres StatefulSet
alongside each AxonHub instance. This is ideal for development, evaluation,
and small deployments.

```yaml
# axonhub-embedded.yaml
apiVersion: axonhub.looplj.com/v1alpha1
kind: AxonHub
metadata:
  name: my-axonhub
  namespace: default
spec:
  image: ghcr.io/looplj/axonhub:latest
  replicas: 1
  port: 8090
  postgres:
    enabled: true
    embedded:
      image: postgres:16
      database: axonhub
      user: axonhub
      storage: 10Gi
      shmSize: 256Mi
      maxConnections: 512
      sharedBuffers: 128MB
```

```sh
kubectl apply -f axonhub-embedded.yaml
```

The operator will:
1. Create a Postgres StatefulSet + Service + Secret with a generated password
2. Wait for Postgres to become healthy
3. Create the AxonHub Deployment with the correct `OCTOPUS_DATABASE_PATH`
4. Update status conditions as resources become ready

### External Postgres (BYO database)

For production, use a managed Postgres (RDS, CloudSQL, CloudNativePG, etc.)
by disabling the embedded database and providing connection details.

First, create a secret with the database password:

```sh
kubectl create secret generic my-pg-secret \
  --from-literal=password='s3cr3t-p@ss'
```

Then create the AxonHub CR:

```yaml
# axonhub-external-db.yaml
apiVersion: axonhub.looplj.com/v1alpha1
kind: AxonHub
metadata:
  name: my-axonhub
  namespace: default
spec:
  image: ghcr.io/looplj/axonhub:latest
  replicas: 1
  port: 8090
  postgres:
    enabled: false
    external:
      host: pg-ha.prod.svc.cluster.local
      port: 5432
      database: axonhub
      user: axonhub
      passwordSecretRef:
        name: my-pg-secret
        key: password
      sslMode: require
```

```sh
kubectl apply -f axonhub-external-db.yaml
```

### Multiple Instances

Each `AxonHub` CR is fully isolated. Create multiple instances in the same
namespace or across namespaces:

```sh
# Instance A in namespace team-a
kubectl create namespace team-a
kubectl apply -n team-a -f axonhub-embedded.yaml

# Instance B in namespace team-b (with external DB)
kubectl create namespace team-b
kubectl apply -n team-b -f axonhub-external-db.yaml
```

```
namespace: team-a
  ├── AxonHub CR "my-axonhub"
  ├── Postgres StatefulSet "my-axonhub-postgres"    ← operator-managed
  ├── Secret "my-axonhub-db-connection"
  └── AxonHub Deployment "my-axonhub"

namespace: team-b
  ├── AxonHub CR "my-axonhub"
  └── AxonHub Deployment "my-axonhub"               ← connects to external DB
```

Check the status:

```sh
kubectl get axonhub -A
# or using the short name
kubectl get ah -A
```

```
NAMESPACE   NAME          READY   REPLICAS   DB READY   AGE
team-a      my-axonhub    true    1          true       2m
team-b      my-axonhub    true    1          true       1m
```

---

## AxonHub CRD Reference

### Spec fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | yes | — | AxonHub container image |
| `replicas` | int32 | no | `1` | Number of AxonHub replicas (1–10) |
| `port` | int32 | no | `8090` | Container listen port |
| `postgres` | object | yes | — | Database configuration |
| `postgres.enabled` | bool | no | `true` | Operator creates a dedicated Postgres |
| `postgres.embedded` | object | no | — | Embedded Postgres config (when enabled) |
| `postgres.external` | object | no | — | External Postgres config (when disabled) |
| `resources` | object | no | — | Compute resources for AxonHub container |
| `env` | array | no | `[]` | Extra environment variables |

### Embedded Postgres fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `image` | string | `postgres:16` | Postgres container image |
| `database` | string | `axonhub` | Database name |
| `user` | string | `axonhub` | Postgres user |
| `passwordSecretRef` | object | _(auto-generated)_ | Secret ref for password |
| `storage` | string | `10Gi` | PVC size |
| `shmSize` | string | `256Mi` | Shared memory size |
| `maxConnections` | int32 | `512` | PostgreSQL `max_connections` |
| `sharedBuffers` | string | `128MB` | PostgreSQL `shared_buffers` |
| `resources` | object | — | Compute resources for Postgres container |

### External Postgres fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `host` | string | yes | — | Postgres host |
| `port` | int32 | no | `5432` | Postgres port |
| `database` | string | no | `axonhub` | Database name |
| `user` | string | no | `axonhub` | Postgres user |
| `passwordSecretRef` | object | yes | — | Secret ref containing password |
| `sslMode` | string | no | `disable` | SSL mode (disable/require/verify-ca/verify-full) |

### Status fields

| Field | Type | Description |
|-------|------|-------------|
| `ready` | bool | AxonHub instance is ready to serve traffic |
| `databaseReady` | bool | Database is ready (embedded or external) |
| `conditions` | array | Detailed condition objects with lastTransitionTime, reason, message |

---

## Metrics & Observability

When `metrics.enabled: true` (default), the operator exposes Prometheus
metrics on port `8443` via the `<release>-metrics-service` service.

**Enabling the ServiceMonitor:**

```sh
helm upgrade axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 1.2.3 \
  --namespace axon-operator-system \
  --set prometheus.monitor.enabled=true \
  --set prometheus.monitor.additionalLabels.release=prometheus
```

**Restricting metrics access with a NetworkPolicy:**

```sh
# Enable the NetworkPolicy
helm upgrade axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version 1.2.3 \
  --namespace axon-operator-system \
  --set networkPolicy.enabled=true

# Label the namespace where Prometheus runs
kubectl label namespace monitoring metrics=enabled
```

---

## Upgrading

To upgrade to a new version:

```sh
helm upgrade axon-operator \
  oci://ghcr.io/amghazanfari/axon-operator \
  --version <new-version> \
  --namespace axon-operator-system \
  -f my-values.yaml
```

Check the rollout status:

```sh
kubectl rollout status deployment/axon-operator-controller-manager \
  -n axon-operator-system
```

---

## Uninstalling

**Remove AxonHub instances first:**

```sh
kubectl delete axonhub --all --all-namespaces
```

**Then uninstall the operator:**

```sh
helm uninstall axon-operator --namespace axon-operator-system
kubectl delete namespace axon-operator-system
```

**Via Make (kustomize install):**

```sh
make undeploy
make uninstall
```

> **Note:** Deleting the `AxonHub` CR triggers cleanup of the Postgres
> StatefulSet, Service, and Secret that the operator created for that
> instance (via owner references). External databases are not touched.

---

## Development

See the root [README.md](../README.md) and [AGENTS.md](../AGENTS.md) for
build, test, and contribution instructions.

### Local development workflow

```sh
# Generate CRDs / RBAC / DeepCopy methods
make manifests generate

# Run the controller locally against the current kubeconfig
make run

# Run unit tests
make test

# Lint
make lint
```

### Releasing

Releases are fully automated via GitHub Actions:

- **Push to `main`** — publishes image tagged `latest` and chart version
  `0.0.0-latest`.
- **Push a tag `vX.Y.Z`** — publishes image and chart with semver tags
  (`X.Y.Z`, `X.Y`, `X`).

```sh
git tag v1.2.3
git push origin v1.2.3
```
