# AGENTS.md

## Project Overview

Kubernetes operator for managing AxonHub deployments, scaffolded with
kubebuilder v4.6.0 (layout: go.kubebuilder.io/v4).

- **Module**: `github.com/amghazanfari/axon-operator`
- **Domain**: `axonhub.looplj.com`
- **Go version**: 1.24.0 (see go.mod)

## Common Commands

### Build

```sh
make build          # Build the manager binary (also runs manifests, generate, fmt, vet)
make docker-build   # Build the Docker image (IMG=controller:latest by default)
```

### Lint

```sh
make lint           # Run golangci-lint (v2.1.0)
make lint-fix       # Run golangci-lint and auto-fix
```

### Test

```sh
make test           # Run unit tests (includes envtest setup)
make test-e2e       # Run e2e tests (requires Kind cluster)
```

### Code Generation

```sh
make manifests      # Generate CRDs, RBAC, and webhook configs
make generate       # Generate DeepCopy methods
```

### Deployment

```sh
make install        # Install CRDs into the cluster
make deploy         # Deploy the controller manager
make undeploy       # Remove the controller manager
make uninstall      # Remove CRDs
```

### Next Step: Create an API

This is a base scaffold (no CRD/controller yet). To add the AxonHub CRD and
controller, run:

```sh
kubebuilder create api --group axonhub --version v1alpha1 --kind AxonHub
```

## CI/CD

GitHub Actions workflows live in `.github/workflows/`:

- `lint.yml` — golangci-lint on push/PR
- `test.yml` — unit tests + Docker image build on push/PR
- `test-e2e.yml` — e2e tests on push to main or manual dispatch
- `release.yml` — builds and pushes multi-arch Docker image to GHCR on tag push

## Environment Notes

The local Go proxy may require `GOSUMDB=off GOPROXY=https://proxy.golang.org,direct`
if the default proxy is unavailable. GitHub Actions CI uses the standard proxy.
