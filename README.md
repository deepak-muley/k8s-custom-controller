# Kubernetes Custom Controller - Resource Counter

A custom Kubernetes controller written in Go that watches and counts instances of specific GroupVersionKind (GVK) resources in a given namespace, including custom resources.

## Features

- **Multi-GVK Support**: Watch and count multiple resource types simultaneously
- **Namespace Filtering**: Count resources in a specific namespace or across all namespaces
- **Custom Resources**: Supports counting custom resource instances
- **Real-time Updates**: Uses Kubernetes informers for efficient, real-time resource tracking
- **Dynamic Discovery**: Automatically discovers resource endpoints via API discovery

## Prerequisites

### Option 1: Devbox (Recommended - Reproducible Environment)

Use Devbox for a fully reproducible development environment with all tools pre-installed:

```bash
# Install devbox (one-time setup)
# macOS: brew install jetpack-io/devbox/devbox
# Linux: curl -fsSL https://get.jetpack.io/devbox | bash

# Start devbox shell (all tools will be available)
make devbox-shell
```

### Option 2: Manual Installation

- Go 1.24 or higher (matches go.mod requirement)
- Access to a Kubernetes cluster (local or remote)
- Valid kubeconfig file (if running outside cluster) or proper RBAC (if running in-cluster)
- **For Makefile targets:**
  - Helm 3.x (for helm chart operations)
  - kubesec (auto-installed via `make kubesec-install`)
  - Pluto (auto-installed via `make pluto-install`)
  - Docker (for Docker builds)
  - Optional: `jq` (for better JSON parsing in kubesec scans)
  - Optional: `golangci-lint` (auto-installed if not present)
- **For Pre-commit hooks:**
  - pre-commit (auto-installed via `make pre-commit-install` or `pip install pre-commit`)
  - Python 3.x (for pre-commit tool)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd k8s-custom-controller
```

2. Install dependencies:
```bash
go mod download
```

3. Build the controller:
```bash
go build -o k8s-controller main.go
```

## Usage

### Basic Usage

Watch and count specific resources in a namespace:

```bash
./k8s-controller \
  -namespace default \
  -gvk "apps/v1/Deployment,apps/v1/ReplicaSet,core/v1/Pod"
```

### Watch Custom Resources

Watch custom resource instances:

```bash
./k8s-controller \
  -namespace production \
  -gvk "example.com/v1/MyCustomResource"
```

### Watch All Namespaces

To watch all namespaces, provide an empty namespace (use quotes to pass empty string):

```bash
./k8s-controller \
  -namespace "" \
  -gvk "apps/v1/Deployment,core/v1/Service"
```

### Using Kubeconfig

Specify a custom kubeconfig file:

```bash
./k8s-controller \
  -kubeconfig ~/.kube/config \
  -namespace default \
  -gvk "apps/v1/Deployment"
```

### Verbose Logging

Increase log verbosity:

```bash
./k8s-controller \
  -v 5 \
  -namespace default \
  -gvk "apps/v1/Deployment"
```

## Command Line Flags

- `-kubeconfig string`: Path to kubeconfig file (optional, defaults to in-cluster config or `~/.kube/config`)
- `-namespace string`: Namespace to watch (empty string for all namespaces, default: "default")
- `-gvk string`: Comma-separated list of GVKs to watch in format `group/version/kind` (required)
- `-v int`: Log level (higher = more verbose, default: 2)

## GVK Format

GVKs should be specified in the format: `group/version/kind`

Examples:
- Core resources: `core/v1/Pod`, `core/v1/Service` (or just `/v1/Pod` for empty group)
- Apps resources: `apps/v1/Deployment`, `apps/v1/ReplicaSet`
- Custom resources: `example.com/v1/MyCustomResource`

## Examples

### Example 1: Count Deployments and Services

```bash
./k8s-controller \
  -namespace default \
  -gvk "apps/v1/Deployment,core/v1/Service"
```

### Example 2: Count Multiple Resource Types

```bash
./k8s-controller \
  -namespace production \
  -gvk "apps/v1/Deployment,apps/v1/StatefulSet,apps/v1/DaemonSet,core/v1/ConfigMap"
```

### Example 3: Watch Custom Resources Only

```bash
./k8s-controller \
  -namespace default \
  -gvk "crd.example.com/v1/MyResource,crd.example.com/v1/AnotherResource"
```

## Output

The controller provides:
- Real-time logs when resources are added, updated, or deleted
- Periodic summary (every 30 seconds) showing current counts for all watched GVKs
- Final summary when the controller shuts down

Example output:
```
Resource Counts:
Last Updated: 2024-01-15T10:30:45Z

  apps/v1/Deployment: 5
  apps/v1/ReplicaSet: 12
  core/v1/Pod: 25
```

## Running as a Deployment

To run the controller as a Kubernetes deployment, you'll need to:

1. Create a Docker image
2. Create a ServiceAccount with appropriate RBAC permissions
3. Deploy as a Deployment or StatefulSet

### Example Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: resource-counter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: resource-counter
  template:
    metadata:
      labels:
        app: resource-counter
    spec:
      serviceAccountName: resource-counter
      containers:
      - name: controller
        image: your-registry/resource-counter:latest
        args:
        - -namespace
        - "default"
        - -gvk
        - "apps/v1/Deployment,apps/v1/ReplicaSet"
```

### Required RBAC Permissions

The controller needs:
- `list`, `watch`, `get` permissions for the resources you want to count
- `list`, `watch` permissions for API discovery

Example ClusterRole:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: resource-counter
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["list", "watch", "get"]
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["list", "watch"]
```

## Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) directory:

- **[Testing Quick Reference](./docs/TESTING_QUICK_REFERENCE.md)** ⭐ **START HERE** - Quick decision guide for when to use each test type
- **[E2E Testing Comparison](./docs/E2E_TESTING_COMPARISON.md)** ⭐ **NEW** - KUTTL vs Golang E2E vs Cluster API Framework
- **[Testing Strategy](./docs/TESTING_STRATEGY.md)** - Complete testing strategy with all options explained
- **[envtest Explained](./docs/ENVTEST_EXPLAINED.md)** - Detailed guide to envtest integration testing
- **[KUTTL FAQ](./docs/KUTTL_FAQ.md)** - Frequently asked questions about KUTTL E2E testing
- **[Setup Guide](./docs/SETUP.md)** - Project setup and installation instructions

See [docs/README.md](./docs/README.md) for the complete documentation index.

## Development

### Project Structure

```
k8s-custom-controller/
├── main.go                          # Entry point
├── main_test.go                     # Main helper function tests
├── go.mod                           # Go module definition
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration management
│   │   └── config_test.go          # Config tests (Ginkgo/Gomega)
│   ├── controller/
│   │   ├── controller.go           # Main controller logic
│   │   ├── controller_test.go      # Controller unit tests (Ginkgo)
│   │   ├── controller_envtest_test.go  # Integration tests (envtest)
│   │   └── suite_test.go           # Test suite setup (envtest)
│   └── counter/
│       ├── counter.go              # Resource counting logic
│       └── counter_test.go         # Counter tests (Ginkgo/Gomega)
├── tests/                           # E2E tests (KUTTL)
│   ├── kuttl-test.yaml             # KUTTL test suite configuration
│   ├── helm-deployment/            # Helm chart deployment tests
│   ├── controller-functionality/   # Controller E2E tests
│   └── namespace-filtering/        # Namespace filtering tests
├── helm/                            # Helm chart
│   └── k8s-custom-controller/
├── docs/                            # Documentation
│   ├── README.md                    # Documentation index
│   ├── TESTING_QUICK_REFERENCE.md   # ⭐ Quick testing decision guide
│   ├── TESTING_STRATEGY.md          # Complete testing strategy
│   ├── ENVTEST_EXPLAINED.md         # envtest detailed guide
│   ├── KUTTL_FAQ.md                 # KUTTL FAQ
│   ├── KUTTL_USAGE.md               # KUTTL usage guide
│   └── ...                          # Additional documentation
└── README.md
```

### Testing Locally

1. Ensure you have access to a Kubernetes cluster:
```bash
kubectl cluster-info
```

2. Run the controller:
```bash
go run main.go -namespace default -gvk "apps/v1/Deployment"
```

## Building with Makefile

The project includes a comprehensive Makefile with targets for building, testing, and deploying.

### Common Makefile Targets

#### Building
```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS)
make build-all

# Build Linux binary only
make build-linux

# Build Docker image
make docker-build

# Build and push Docker image
make docker-push
```

#### Testing
```bash
# Run all tests (unit + integration with ginkgo/envtest)
make test

# Run tests with coverage report
make test-coverage

# Run unit tests only
make test-unit

# Run integration tests with envtest
make test-integration

# Run E2E tests with KUTTL (requires cluster or kind) - Optional
make kuttl-install      # Install KUTTL first
make kuttl-test         # Run KUTTL tests (requires cluster)
make kuttl-test-kind    # Run with kind cluster (auto-creates/destroys)
make kuttl-test-helm    # Run Helm chart tests only
make kuttl-clean        # Clean up test namespaces

# See docs/TESTING_QUICK_REFERENCE.md for when to use each test type
```

#### Code Quality
```bash
# Format code
make fmt

# Run go vet
make vet

# Run linter (golangci-lint)
make lint

# Tidy go modules
make tidy
```

#### Helm Chart Operations
```bash
# Create Helm chart structure
make helm-create

# Lint Helm chart
make helm-lint

# Render Helm templates to stdout
make helm-template

# Render Helm templates to directory
make helm-template-dir

# Package Helm chart
make helm-package

# Build and package Helm chart
make helm-build
```

#### Security Scanning with Kubesec
```bash
# Install kubesec
make kubesec-install

# Scan rendered Helm templates (default min score: 90)
make kubesec-scan

# Scan and generate JSON reports
make kubesec-scan-json

# Validate all templates meet minimum score
make kubesec-validate

# Run full security scan (render + validate)
make security-scan
```

#### API Version Validation with Pluto
```bash
# Install Pluto (detects deprecated Kubernetes apiVersions)
make pluto-install

# Detect deprecated apiVersions in rendered Helm templates
make pluto-detect-files

# Detect deprecated apiVersions in Helm chart templates
make pluto-detect-charts

# Validate no deprecated apiVersions (fails on deprecations)
make pluto-detect-validate PLUTO_TARGET_K8S_VERSION=1.29

# Detect deprecated apiVersions in both charts and rendered templates
make pluto-detect-all

# Custom target Kubernetes version
make pluto-detect-validate PLUTO_TARGET_K8S_VERSION=1.30
```

[Pluto](https://github.com/FairwindsOps/pluto) helps detect deprecated and removed Kubernetes apiVersions in your Helm charts and YAML manifests, preventing API breakage issues during Kubernetes upgrades.

#### CI Pipeline
```bash
# Run full CI pipeline (format, vet, lint, test, build, helm-lint, security-scan)
make ci
```

#### Devbox Shell (Reproducible Development Environment)
```bash
# Check if devbox is installed (shows installation instructions if not)
make devbox-install

# Start devbox shell (all tools pre-installed)
make devbox-shell

# Inside devbox shell, most tools are available:
# - Go 1.24+, kubectl, jq, git, curl, bash, python3
# - Helm (may need manual installation if package not available)
# - Run any make commands as usual (build, test, ci, etc.)

# Run a command in devbox without entering shell
make devbox-run CMD="make build"
make devbox-run CMD="make test"
make devbox-run CMD="make ci"

# Initialize devbox (if devbox.json is missing - should already exist)
make devbox-init

# Update devbox packages to latest versions
make devbox-update

# Show devbox version and installed packages
make devbox-info
```

**Devbox benefits:**
- ✅ All development tools pre-installed (Go, Helm, kubectl, jq, etc.)
- ✅ Reproducible environment across different machines (macOS, Linux, CI)
- ✅ No need to manually install dependencies
- ✅ Works the same on macOS, Linux, and CI environments
- ✅ Isolated from system packages (no conflicts)

**To exit devbox shell:**
```bash
exit
# or press Ctrl+D
```

**Troubleshooting:**

If you encounter package errors (e.g., `helm@3.12.0: package not found`), try:

1. **Update devbox packages:**
   ```bash
   make devbox-update
   ```

2. **If Helm is not available in devbox:**
   - Install Helm manually inside the devbox shell:
     ```bash
     # macOS (inside devbox shell)
     brew install helm
     
     # Linux (inside devbox shell)
     curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
     ```

3. **Install additional tools:**
   Tools like `kubesec` and `pluto` need to be installed separately as they are not available in nixpkgs:
   ```bash
   # Inside devbox shell or via devbox script
   devbox run make install-tools
   # or manually
   make kubesec-install
   make pluto-install
   ```

**Note:** Some packages may not be available in devbox/nixpkgs. If a package is not found, install it manually using your system package manager (brew on macOS, apt/yum on Linux) inside the devbox shell.

#### Pre-commit Hooks
```bash
# Install pre-commit hooks (run once after cloning)
# This installs pre-commit tool and sets up git hooks
make pre-commit-install

# Run pre-commit hooks on all files (before committing)
make pre-commit-run

# Or let pre-commit run automatically on git commit
# (hooks run automatically after installation)

# Update pre-commit hooks to latest versions
make pre-commit-update

# Uninstall pre-commit hooks
make pre-commit-uninstall

# Run pre-push checks (lighter than full CI, no security scans)
make pre-push
```

**Pre-commit hooks run automatically** on `git commit` after installation. Install hooks once with `make pre-commit-install`.

**Pre-commit checks include:**
- Code formatting (`make fmt`)
- Go vet (`make vet`)
- Unit tests (`make test-unit`)
- Linting (golangci-lint)
- YAML/JSON validation
- Helm chart linting (on Helm files)
- General file checks (trailing whitespace, end-of-file, etc.)

**To skip pre-commit hooks** (not recommended):
```bash
git commit --no-verify -m "message"
```

#### Cleanup
```bash
# Clean build artifacts
make clean

# Clean all artifacts including helm chart
make clean-all
```

### Customizing Makefile Variables

You can override default values:

```bash
# Set minimum kubesec score
make kubesec-validate KUBESEC_MIN_SCORE=95

# Set namespace for helm templates
make helm-template NAMESPACE=production

# Set image repository and tag
make docker-build IMAGE_REPO=myregistry/k8s-controller IMAGE_TAG=v1.0.0
```

### Full Example: Build and Security Scan

```bash
# 1. Build the project
make build

# 2. Run tests
make test

# 3. Build Docker image
make docker-build IMAGE_TAG=v1.0.0

# 4. Build Helm chart
make helm-build

# 5. Render templates and validate with kubesec
make security-scan KUBESEC_MIN_SCORE=90

# 6. Or run full CI pipeline
make ci
```

## Security Features

The Helm chart includes security best practices for high kubesec scores:

- **Non-root user**: Runs as UID 65534 (nobody)
- **Read-only root filesystem**: Container filesystem is read-only with tmp volume
- **No privilege escalation**: `allowPrivilegeEscalation: false`
- **Dropped capabilities**: All Linux capabilities dropped
- **Security contexts**: Pod and container security contexts configured
- **Seccomp profile**: Runtime default seccomp profile
- **Resource limits**: CPU and memory limits defined
- **Service account**: Dedicated service account with minimal permissions
- **RBAC**: Role-based access control with least privilege

These configurations ensure the deployment achieves a kubesec score of 90+ out of 100.

## Limitations

- The controller uses a simple counting mechanism and doesn't track individual resource instances
- Resource discovery relies on the Kubernetes API server's discovery endpoints
- Very large numbers of resources may impact performance
- Health probes require HTTP endpoints to be implemented in the controller

## Controller-Runtime Examples

This repository includes comprehensive examples demonstrating controller-runtime patterns alongside the client-go based main controller. These examples showcase modern Kubernetes controller development with CRDs, webhooks, finalizers, and more.

### Quick Start with Examples

```bash
# List all available examples
make examples-list

# Generate CRDs for all examples
make examples-generate

# Build all example controllers
make examples-build

# Run a specific example
make example-run EXAMPLE=01-basic-reconciler

# Install CRDs to your cluster
make examples-install-crds

# Run tests for all examples
make examples-test
```

### Available Examples

The examples progress from basic concepts to production patterns:

**Foundation:**
- [01-basic-reconciler](examples/01-basic-reconciler) - Manager setup, Reconcile loop, CRD basics
- [02-predicates-filtering](examples/02-predicates-filtering) - Event filtering, performance optimization
- [03-finalizers-cleanup](examples/03-finalizers-cleanup) 🔥 - Cleanup logic, external resource management

**Advanced Patterns:**
- [04-owner-references](examples/04-owner-references) - Resource ownership, automatic garbage collection
- [05-status-conditions](examples/05-status-conditions) 🔥 - Status management, condition patterns
- [06-multi-resource-watch](examples/06-multi-resource-watch) - Watching multiple resource types

**Production Features:**
- [07-webhooks](examples/07-webhooks) 🔥 - Validation & mutating webhooks
- [08-metrics-events](examples/08-metrics-events) - Prometheus metrics, event recording
- [09-advanced-indexing](examples/09-advanced-indexing) - Cache optimization, fast lookups
- [10-event-source-chaining](examples/10-event-source-chaining) ⭐ - External event sources
- [11-rate-limiting-backoff](examples/11-rate-limiting-backoff) ⭐ - Rate limiting, retry strategies

### Documentation

- **[Examples Overview](examples/README.md)** - Complete guide to all examples
- **[Controller-Runtime Guide](docs/CONTROLLER_RUNTIME_GUIDE.md)** - When to use controller-runtime vs client-go
- **[CRD Development Guide](docs/CRD_DEVELOPMENT.md)** - Kubebuilder markers and CRD patterns
- **[Migration Guide](docs/MIGRATION_GUIDE.md)** - Converting client-go to controller-runtime

### Learning Path

1. **Study the main controller** (client-go approach) - `main.go`, `internal/controller/`
2. **Compare with example 01** (controller-runtime approach)
3. **Work through examples 03, 05, 07** (priority features)
4. **Explore advanced examples** as needed

See [examples/README.md](examples/README.md) for detailed learning paths and [docs/CONTROLLER_RUNTIME_GUIDE.md](docs/CONTROLLER_RUNTIME_GUIDE.md) for choosing the right approach.

## License

[Your License Here]

## GitHub Actions CI/CD

The repository includes a comprehensive GitHub Actions workflow (`.github/workflows/ci.yml`) that automates:

### CI Pipeline Jobs

1. **Test and Lint** - Runs tests, code formatting, vet, and linter
2. **Build** - Builds binaries for all platforms
3. **Docker Build** - Builds and pushes Docker images to GHCR
4. **Helm Chart** - Builds, lints, validates with kubesec and Pluto, and packages Helm chart
5. **Push Helm Chart** - Pushes Helm chart as OCI artifact to GHCR
6. **CI Summary** - Generates summary of all pipeline results

### Workflow Triggers

- **Push** to `main`, `master`, or `develop` branches
- **Pull requests** to `main`, `master`, or `develop` branches
- **Tags** matching `v*` pattern (e.g., `v1.0.0`)

### Features

- **Kubesec Scanning**: Automatically scans all Helm templates with minimum score of 90
- **Pluto API Version Validation**: Detects deprecated and removed Kubernetes apiVersions to prevent API breakage (target: K8s 1.29)
- **OCI Artifact Push**: Pushes Helm charts to `ghcr.io/deepak-muley/k8s-custom-controller-chart` with version tags
- **Docker Image Push**: Uses Makefile to build and push Docker images (multi-platform support)
- **Multi-platform Build**: Builds for both amd64 and arm64 architectures
- **Coverage Reports**: Uploads test coverage to codecov
- **Makefile Integration**: All commands use Makefile targets only

### Using the Helm Chart from GHCR

Once the chart is pushed to GHCR, you can use it:

```bash
# Authenticate to GHCR
helm registry login ghcr.io -u USERNAME -p TOKEN

# Note: Helm chart has -chart postfix with version tag
# Pull the chart with version
helm pull oci://ghcr.io/deepak-muley/k8s-custom-controller-chart:1.0.0

# Or install directly (note: chart has -chart postfix)
helm install my-controller oci://ghcr.io/deepak-muley/k8s-custom-controller-chart:1.0.0 \
  --namespace default \
  --set controller.namespace=default \
  --set controller.gvks="apps/v1/Deployment,apps/v1/ReplicaSet"
```

**Note**: The Helm chart is pushed with a `-chart` postfix and explicit version tag:
- Chart location: `oci://ghcr.io/deepak-muley/k8s-custom-controller-chart:version`
- Docker image: `ghcr.io/deepak-muley/k8s-custom-controller:version`

### Manual Trigger

The workflow can also be triggered manually from the GitHub Actions UI.

## Contributing

[Contributing guidelines]

