# Kubernetes Custom Controller - Resource Counter

A custom Kubernetes controller written in Go that watches and counts instances of specific GroupVersionKind (GVK) resources in a given namespace, including custom resources.

## Features

- **Multi-GVK Support**: Watch and count multiple resource types simultaneously
- **Namespace Filtering**: Count resources in a specific namespace or across all namespaces
- **Custom Resources**: Supports counting custom resource instances
- **Real-time Updates**: Uses Kubernetes informers for efficient, real-time resource tracking
- **Dynamic Discovery**: Automatically discovers resource endpoints via API discovery

## Prerequisites

- Go 1.21 or higher
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

## Development

### Project Structure

```
k8s-custom-controller/
├── main.go                          # Entry point
├── go.mod                           # Go module definition
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── controller/
│   │   └── controller.go           # Main controller logic
│   └── counter/
│       └── counter.go              # Resource counting logic
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
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run unit tests only
make test-unit
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

