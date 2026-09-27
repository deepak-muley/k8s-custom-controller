# KUTTL E2E Tests

This directory contains end-to-end (E2E) tests using [KUTTL](https://github.com/kudobuilder/kuttl) (Kubernetes Test TooL).

## What is KUTTL?

KUTTL is a **declarative testing framework** for Kubernetes that uses YAML manifests to test Kubernetes resources in a real cluster. It's perfect for:

- ✅ **Helm Chart Testing** - Test complete Helm chart deployments (full install)
- ✅ **E2E Testing** - Test end-to-end workflows
- ✅ **Operator Testing** - Test Kubernetes operators
- ✅ **Integration Testing** - Test multi-component interactions

Reference: [KUTTL GitHub](https://github.com/kudobuilder/kuttl)

## Test Structure

```
tests/
├── kuttl-test.yaml              # KUTTL test suite configuration
│
├── helm-deployment/             # Test Helm chart deployment (FULL HELM INSTALL)
│   ├── 00-install-helm.sh      # ⭐ Script that installs ENTIRE Helm chart
│   ├── 00-install.yaml         # Setup namespace
│   ├── 01-assert-resources.yaml # Verify all resources created by Helm chart
│   └── steps/                   # Additional verification steps
│       ├── 01-verify-controller-functionality.yaml
│       ├── 02-assert-deployments-ready.yaml
│       └── 03-verify-controller-counting.yaml
│
├── controller-functionality/    # Test controller behavior
│   ├── 00-setup.yaml           # Setup test environment
│   └── steps/
│       ├── 00-create-test-deployments.yaml
│       ├── 01-assert-deployments-exist.yaml
│       └── 02-verify-controller-counts.yaml
│
└── namespace-filtering/         # Test namespace filtering
    ├── 00-setup.yaml           # Create test namespaces
    └── steps/
        ├── 00-create-deployments-both-namespaces.yaml
        └── 01-assert-only-watched-counted.yaml
```

## Helm Chart Installation Example ⭐

### Full Helm Chart Install (Not Individual YAMLs)

The `helm-deployment` test suite demonstrates installing the **entire Helm chart** using `helm install` (not individual YAMLs):

#### 1. Installation Script (`00-install-helm.sh`)

```bash
#!/bin/bash
# This script performs a FULL Helm chart installation
# It installs ALL resources defined in the Helm chart via "helm install"

helm upgrade --install k8s-custom-controller ./helm/k8s-custom-controller \
  --namespace kuttl-test-helm \
  --set image.repository=k8s-custom-controller \
  --set image.tag=test \
  --set controller.namespace=kuttl-test-controller \
  --set controller.gvks="apps/v1/Deployment" \
  --wait \
  --timeout 5m
```

**What this installs (all via Helm):**
- ✅ Deployment (with all templates and values)
- ✅ ServiceAccount
- ✅ ClusterRole (with RBAC rules)
- ✅ ClusterRoleBinding
- ✅ ConfigMap (if configured)
- ✅ NetworkPolicy (if configured)
- ✅ All other resources defined in Helm templates

**Key difference from individual YAMLs:**
- ❌ **NOT** installing individual YAML files
- ✅ **FULL** Helm chart installation with all templates, values, and hooks

#### 2. Resource Verification (`01-assert-resources.yaml`)

After Helm installs the chart, KUTTL verifies all resources exist:

```yaml
# Verify Deployment (created by Helm chart)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1  # KUTTL waits for this state

# Verify ServiceAccount (created by Helm chart)
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm

# Verify ClusterRole (created by Helm chart)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-custom-controller

# Verify ClusterRoleBinding (created by Helm chart)
# ... etc
```

#### 3. Functionality Testing (`steps/`)

Additional steps verify the controller works after Helm installation:

- Create test resources
- Verify controller counts resources
- Test complete workflow

## Running Tests

### Prerequisites

1. **KUTTL installed** - See [KUTTL Usage Guide](../docs/KUTTL_USAGE.md#installing-kuttl)
2. **Helm installed** - Required for Helm chart installation tests
3. **Kubernetes cluster** - kind, minikube, or existing cluster
4. **kubectl configured** - Access to test cluster
5. **Docker image built** - For controller deployment tests

### Quick Start

```bash
# 1. Build Docker image
make docker-build IMAGE_TAG=test

# 2. Package Helm chart
make helm-package

# 3. Run with kind cluster (auto-creates, installs Helm chart, runs tests)
make kuttl-test-kind

# This will:
# - Create kind cluster
# - Build and load Docker image into kind
# - Package Helm chart
# - Run KUTTL tests (installs Helm chart via 00-install-helm.sh script)
# - Verify all resources created by Helm chart
# - Clean up kind cluster
```

### Manual Usage

```bash
# 1. Prepare: Build Docker image and package Helm chart
make docker-build IMAGE_TAG=test
make helm-package

# 2. If using kind, create cluster and load image
kind create cluster --name kuttl-test
kind load docker-image k8s-custom-controller:test --name kuttl-test

# 3. Run Helm chart deployment tests
# The 00-install-helm.sh script will automatically:
# - Install complete Helm chart via "helm install"
# - Install ALL resources (Deployment, SA, RBAC, ConfigMap, etc.)
# - Wait for deployment to be ready
kubectl kuttl test tests/helm-deployment/ --start-kind=false

# 4. Cleanup
kind delete cluster --name kuttl-test
```

### Basic Usage

```bash
# Run all tests (includes Helm chart installation)
kubectl kuttl test tests/

# Run specific test suite (Helm chart deployment)
kubectl kuttl test tests/helm-deployment/

# Run with verbose output
kubectl kuttl test tests/ --debug

# Use existing cluster (don't create kind)
kubectl kuttl test tests/ --start-kind=false

# Cleanup namespaces after success
kubectl kuttl test tests/ --delete-namespace-on-success
```

## Where to Use KUTTL

### ✅ 1. Helm Chart Testing (Primary Use Case)

**KUTTL is excellent for testing Helm charts** because it can:

1. **Deploy Helm Chart** - Test complete Helm chart installation (full `helm install`)
2. **Verify Resources** - Ensure all resources (Deployment, ServiceAccount, RBAC, ConfigMap, etc.) are created correctly
3. **Test Configurations** - Test with different values.yaml configurations
4. **Upgrade Testing** - Test Helm chart upgrades between versions
5. **Uninstall Testing** - Verify clean uninstallation

**Example:**
```yaml
# The 00-install-helm.sh script installs the entire chart:
# helm install k8s-custom-controller ./helm/k8s-custom-controller \
#   --set controller.namespace=test \
#   --set controller.gvks="apps/v1/Deployment"
#
# Then 01-assert-resources.yaml verifies all resources exist
```

### ✅ 2. End-to-End Controller Testing

**Test complete controller workflows:**

- Deploy controller via Helm → Create test resources → Verify counting
- Test namespace filtering
- Test resource creation/deletion workflows
- Test controller restart and recovery

### ✅ 3. Integration Testing

**Test interactions between components:**

- Controller + other Kubernetes resources
- Multi-namespace scenarios
- RBAC and security policies
- Complete workflows from deployment to functionality

## KUTTL vs Other Testing Tools

| Tool | Purpose | Speed | Cluster Needed | Best For |
|------|---------|-------|----------------|----------|
| **Ginkgo/Gomega** | Unit tests (Go code) | ⚡ Fast | ❌ No | Counter logic, parsing |
| **envtest** | Integration tests | ⚡ Fast (1-3s) | ❌ No (API server in-process) | Controller logic, informers |
| **KUTTL** | Helm charts, E2E | ⚠️ Slower (10-30s) | ✅ Yes (kind) | Helm deployment, E2E workflows |
| **Golang E2E** | Complex E2E (code-based) | ⚠️ Slower (10-30s) | ✅ Yes (kind) | Complex scenarios, full control |
| **Cluster API Framework** | Large projects (framework) | ❌ Slow (setup) | ✅ Yes | Large projects, reusable patterns |

**See [E2E Testing Comparison](../docs/E2E_TESTING_COMPARISON.md) for detailed comparison.**

## When to Use KUTTL

### ✅ Use KUTTL for:

1. **Helm Chart E2E Testing** ⭐ **Primary Use Case**
   - Deploy Helm chart → Verify resources → Test functionality
   - Test with different values.yaml configurations
   - Test upgrade/downgrade scenarios

2. **Simple E2E Workflows**
   - Complete deployment workflows
   - Resource creation and validation
   - Quick feedback in CI/CD

### ❌ Don't Use KUTTL for:

1. **Unit Testing** - Use Ginkgo/Gomega (already implemented)
2. **Controller Logic Testing** - Use envtest (already implemented)
3. **Fast Feedback** - Use unit/integration tests (faster)

## Example: Complete Helm Chart Test Workflow

### Test Scenario: Helm Chart Deployment

```
1. BeforeSuite (kuttl-test.yaml)
   └─> Create kind cluster (if startKind=true)
   
2. Test Suite: helm-deployment/
   ├─> 00-install-helm.sh
   │   └─> Run: helm install k8s-custom-controller ./helm/k8s-custom-controller
   │       └─> Installs ALL resources: Deployment, SA, RBAC, ConfigMap, etc.
   │
   └─> 01-assert-resources.yaml
       ├─> Wait for Deployment to exist (readyReplicas: 1)
       ├─> Wait for ServiceAccount to exist
       ├─> Wait for RBAC to exist
       └─> Verify all resources are in expected state

3. Steps: helm-deployment/steps/
   ├─> 01-verify-controller-functionality.yaml
   │   └─> Create test resources, verify controller works
   │
   ├─> 02-assert-deployments-ready.yaml
   │   └─> Verify test resources are ready
   │
   └─> 03-verify-controller-counting.yaml
       └─> Verify controller is counting resources

4. AfterSuite (kuttl-test.yaml)
   └─> Delete kind cluster (if created)
```

### Running the Complete Test

```bash
# One command does everything:
make kuttl-test-kind

# This will:
# 1. Create kind cluster
# 2. Build Docker image
# 3. Load image into kind
# 4. Package Helm chart
# 5. Run KUTTL tests:
#    - Execute 00-install-helm.sh (installs full Helm chart)
#    - Verify all resources in 01-assert-resources.yaml
#    - Run additional verification steps
# 6. Clean up kind cluster
```

## Important Notes

### 1. Helm Chart Installation

**✅ Full Helm Install (What We Use):**
- Uses `helm install` command in script
- Installs entire chart with all templates and values
- Respects Helm hooks, dependencies, and lifecycle
- Tests real Helm deployment workflow

**❌ NOT Individual YAMLs:**
- Not installing individual YAML files manually
- Not testing rendered templates only
- Not testing static YAML files

### 2. Controller State Verification

**Challenge:** KUTTL can verify resources exist, but verifying controller's internal count state is harder.

**Solutions:**
1. **Add Metrics Endpoint** - Expose controller counts via Prometheus metrics
2. **Add Health Endpoint** - Create `/health` endpoint that returns counts
3. **Check Logs** - Query controller logs for count information
4. **Test Pod** - Use a test pod to exec into controller and query state

### 3. Test Cluster Setup

**Options:**
1. **kind** - Fast, lightweight, CI/CD friendly (recommended)
2. **minikube** - Local testing
3. **Existing Cluster** - Use current kubeconfig context
4. **KUTTL Auto-Managed** - Let KUTTL create kind cluster (startKind=true)

## Comparison with Other E2E Approaches

See **[E2E Testing Comparison](../docs/E2E_TESTING_COMPARISON.md)** for detailed comparison between:
- **KUTTL** (YAML-based, what we use)
- **Golang E2E with kind** (Code-based)
- **Cluster API E2E Framework** (Framework-based)

**Quick Summary:**
- **KUTTL** - Best for Helm chart testing, simple E2E workflows
- **Golang E2E** - Best for complex scenarios requiring full control
- **Cluster API Framework** - Best for large projects, but overkill for simple controllers

## References

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl)
- [KUTTL Documentation](https://kuttl.dev/)
- [E2E Testing Comparison](../docs/E2E_TESTING_COMPARISON.md) - ⭐ KUTTL vs Golang E2E vs Cluster API Framework
- [Testing Quick Reference](../docs/TESTING_QUICK_REFERENCE.md) - Quick decision guide
- [Testing Strategy](../docs/TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL FAQ](../docs/KUTTL_FAQ.md) - Frequently asked questions
- [KUTTL Usage](../docs/KUTTL_USAGE.md) - Detailed usage guide
