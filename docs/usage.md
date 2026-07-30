# AxonHub Operator — Usage Guide

This document describes how to install, configure, and use the AxonHub
Operator in a Kubernetes cluster.

> **Status:** The operator is currently in the scaffolding phase. CRD types
> and controller reconciliation logic will be added in subsequent iterations.
> Sections marked with _(pending)_ are placeholders for future functionality.

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
- [Deploying AxonHub Instances _(pending)_](#deploying-axonhub-instances-pending)
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

Verify RBAC resources were created:

```sh
kubectl get clusterrole | grep axon-operator
kubectl get clusterrolebinding | grep axon-operator
```

---

## Deploying AxonHub Instances _(pending)_

> **Note:** The `AxonHub` CRD does not exist yet. This section is a scaffold
> and will be populated once the API types and controller are implemented.

Once the CRD is available, you will be able to create an AxonHub instance:

```yaml
# axonhub-sample.yaml
apiVersion: axonhub.looplj.com/v1alpha1
kind: AxonHub
metadata:
  name: my-axonhub
  namespace: default
spec:
  # Image of the AxonHub server to deploy.
  image: ghcr.io/looplj/axonhub:latest
  # Number of replicas.
  replicas: 1
  # Service type for exposing the AxonHub API.
  serviceType: ClusterIP
  # TODO: add configuration fields (port, resources, env, etc.)
```

```sh
kubectl apply -f axonhub-sample.yaml
```

Check the status:

```sh
kubectl get axonhub
kubectl describe axonhub my-axonhub
```

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

**Via Helm:**

```sh
helm uninstall axon-operator --namespace axon-operator-system
kubectl delete namespace axon-operator-system
```

**Via Make (kustomize install):**

```sh
make undeploy
make uninstall
```

> **Note:** Removing the operator does not delete the AxonHub instances (CRs)
> you may have created. Remove those first if you want a full cleanup.

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
