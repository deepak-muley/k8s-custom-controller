# KUTTL Usage Guide for This Project

## Overview

This document explains **where and how to use KUTTL** in this project, particularly for Helm chart testing and end-to-end testing.

Reference: [KUTTL GitHub Repository](https://github.com/kudobuilder/kuttl)

## Where to Use KUTTL

### 1. ✅ Helm Chart Testing (Recommended)

**KUTTL is perfect for testing your Helm chart** because it can:

1. **Deploy Helm Chart** - Test that your chart installs correctly
2. **Verify Resources** - Ensure all resources (Deployment, ServiceAccount, RBAC, etc.) are created
3. **Test Configurations** - Test with different values.yaml configurations
4. **Upgrade Testing** - Test Helm chart upgrades between versions
5. **Uninstall Testing** - Verify clean uninstallation

**Example Use Cases:**
- Test Helm chart deployment in a real cluster
- Verify all Kubernetes resources are created correctly
- Test with different namespace configurations
- Test upgrade from v1.0.0 to v1.1.0
- Verify RBAC permissions are correct
- Test security contexts are applied correctly

### 2. ✅ End-to-End Controller Testing

**Test complete controller workflows:**

1. **Deploy Controller** - Install controller via Helm chart
2. **Create Resources** - Create Deployments/Services for controller to count
3. **Verify Counting** - Verify controller correctly counts resources
4. **Test Namespace Filtering** - Verify namespace filtering works
5. **Test Resource Deletion** - Verify count decreases when resources deleted

**Example Use Cases:**
- Deploy controller → Create deployments → Verify counting works
- Test namespace filtering (watch one namespace, ignore others)
- Test multiple GVKs simultaneously
- Test controller restart and recovery

### 3. ✅ Integration Testing

**Test interactions between components:**

1. **Controller + Resources** - Test controller with actual Kubernetes resources
2. **Multi-Namespace** - Test cross-namespace functionality
3. **RBAC** - Verify service account and RBAC work correctly
4. **Network Policies** - Test network policy enforcement (if configured)
5. **ConfigMaps/Secrets** - Verify configuration is loaded correctly

## KUTTL vs Other Testing Tools in Your Project

| Tool | Purpose | File Location | When to Use |
|------|---------|---------------|-------------|
| **Ginkgo/Gomega** | Unit tests (counter, config) | `internal/*/*_test.go` | Test Go code logic |
| **envtest** | Controller integration tests | `internal/controller/*_envtest_test.go` | Test controller with real API server |
| **KUTTL** | Helm chart E2E tests | `tests/*/` | Test Helm deployment and E2E workflows |
| **kubesec** | Security scanning | Makefile | Scan rendered Helm templates |
| **Pluto** | API version validation | Makefile | Check for deprecated API versions |
| **helm-unittest** | Template validation | (Not implemented) | Test Helm template rendering |

## Test Structure

```
tests/
├── kuttl-test.yaml              # KUTTL test suite configuration
│
├── helm-deployment/             # Test Helm chart deployment
│   ├── 00-install.yaml         # Install Helm chart (via script)
│   └── 01-assert-resources.yaml # Verify resources exist
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

## Installing KUTTL

### Option 1: Download Binary (Recommended)

```bash
# Download latest release
curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_$(uname -s)_$(uname -m) -o kubectl-kuttl
chmod +x kubectl-kuttl
sudo mv kubectl-kuttl /usr/local/bin/kubectl-kuttl

# Verify installation
kubectl kuttl version
```

### Option 2: kubectl Plugin (via krew)

```bash
# Install krew first: https://krew.sigs.k8s.io/docs/user-guide/setup/install/
kubectl krew install kuttl

# Use as kubectl plugin
kubectl kuttl test tests/
```

### Option 3: Homebrew (macOS)

```bash
brew install kuttl
```

### Option 4: Add to Devbox

You can add KUTTL to your `devbox.json`:

```json
{
  "packages": [
    "go@1.24",
    "kubectl",
    "helm",
    "kuttl"  // Add this if available in nixpkgs
  ]
}
```

Or install manually in devbox:

```bash
# Inside devbox shell
curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_$(uname -s)_$(uname -m) -o kubectl-kuttl
chmod +x kubectl-kuttl
mv kubectl-kuttl ~/.local/bin/kubectl-kuttl
```

## Running KUTTL Tests

### Basic Usage

```bash
# Run all tests in tests/ directory
kubectl kuttl test tests/

# Run specific test suite
kubectl kuttl test tests/helm-deployment

# Run with verbose output
kubectl kuttl test tests/ --debug

# Run without creating kind cluster (use existing kubeconfig)
kubectl kuttl test tests/ --start-kind=false

# Run with namespace cleanup
kubectl kuttl test tests/ --delete-namespace-on-success
```

### Using Test Configuration File

```bash
# Use kuttl-test.yaml configuration
kubectl kuttl test --config tests/kuttl-test.yaml

# Or specify test directory
kubectl kuttl test tests/ --config tests/kuttl-test.yaml
```

### CI/CD Integration

```bash
# In GitHub Actions or other CI
# Option 1: Use existing cluster
kubectl kuttl test tests/ --start-kind=false

# Option 2: Create kind cluster
kubectl kuttl test tests/ --start-kind=true

# Option 3: Use kind separately (for faster parallel tests)
kind create cluster --name kuttl-test
kubectl kuttl test tests/ --start-kind=false
kind delete cluster --name kuttl-test
```

## Example: Complete KUTTL Test Workflow

### Test Scenario: Helm Chart Deployment

```bash
# 1. Install KUTTL
curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_$(uname -s)_$(uname -m) -o kubectl-kuttl
chmod +x kubectl-kuttl
sudo mv kubectl-kuttl /usr/local/bin/kubectl-kuttl

# 2. Create or use a test cluster
# Option A: Use kind
kind create cluster --name kuttl-test

# Option B: Use existing cluster
# kubectl config use-context your-test-cluster

# 3. Build and push your Helm chart
make helm-package
helm install test-controller ./helm/k8s-custom-controller/charts/k8s-custom-controller-*.tgz \
  --namespace kuttl-test-helm \
  --create-namespace \
  --set controller.namespace=kuttl-test-controller \
  --set controller.gvks="apps/v1/Deployment"

# 4. Run KUTTL tests
kubectl kuttl test tests/controller-functionality \
  --start-kind=false \
  --delete-namespace-on-success

# 5. Cleanup (if needed)
kind delete cluster --name kuttl-test
```

## Limitations and Workarounds

### Limitation 1: Controller State Verification

**Problem:** KUTTL can verify resources exist, but verifying controller's internal count state is harder.

**Solution:**
1. **Add Metrics Endpoint** - Expose controller counts via Prometheus metrics
2. **Add Health Endpoint** - Create a `/health` endpoint that returns counts
3. **Check Logs** - Query controller logs for count information
4. **Test Pod** - Use a test pod to exec into controller and query state

### Limitation 2: Helm Chart Installation

**Problem:** KUTTL doesn't directly install Helm charts from YAML files.

**Solution:**
1. **Use Scripts** - Create scripts that install Helm charts before tests
2. **Test Rendered Templates** - Render Helm templates and test the YAML output
3. **Pre-Deploy Hooks** - Use KUTTL's ability to run commands before tests

### Limitation 3: Controller Startup Time

**Problem:** Controller needs time to start and sync informers before counting.

**Solution:**
1. **Add Wait Conditions** - Use KUTTL's assertion capabilities to wait for ready state
2. **Retry Logic** - KUTTL automatically retries assertions until timeout
3. **Health Checks** - Add readiness/liveness probes to controller deployment

## Integration with CI/CD

### GitHub Actions Example

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests with KUTTL

on:
  pull_request:
    branches: [main, master]
  push:
    branches: [main, master]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Install KUTTL
        run: |
          curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_linux_amd64 -o kubectl-kuttl
          chmod +x kubectl-kuttl
          sudo mv kubectl-kuttl /usr/local/bin/kubectl-kuttl
      
      - name: Create kind cluster
        uses: helm/kind-action@v1.8.0
        with:
          node_image: kindest/node:v1.28.0
      
      - name: Build Docker image
        run: |
          docker build -t k8s-custom-controller:test .
          kind load docker-image k8s-custom-controller:test
      
      - name: Package Helm chart
        run: make helm-package
      
      - name: Install Helm chart
        run: |
          helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
            --namespace kuttl-test-helm \
            --create-namespace \
            --set image.repository=k8s-custom-controller \
            --set image.tag=test \
            --set controller.namespace=kuttl-test-controller \
            --set controller.gvks="apps/v1/Deployment"
      
      - name: Run KUTTL tests
        run: kubectl kuttl test tests/ --start-kind=false
      
      - name: Cleanup
        if: always()
        run: kind delete cluster
```

## Best Practices

1. **Use Separate Namespaces** - Create unique test namespaces for isolation
2. **Cleanup Resources** - Always clean up test namespaces after tests
3. **Use Assertions** - Verify expected state, not just resource existence
4. **Test Multiple Scenarios** - Test happy path, error cases, edge cases
5. **Keep Tests Fast** - Use parallel execution where possible
6. **Document Test Purpose** - Each test directory should have clear purpose
7. **Version Control** - Commit test YAML files to version control
8. **CI Integration** - Run KUTTL tests in CI/CD pipeline

## Summary: When to Use Each Testing Approach

### Use Ginkgo/Gomega for:
- ✅ Unit testing Go code (counter, config, parsing)
- ✅ Fast feedback during development
- ✅ Testing isolated business logic

### Use envtest for:
- ✅ Testing controller logic with real API server
- ✅ Testing informers and cache syncing
- ✅ Testing discovery API interactions
- ✅ Integration testing without full cluster

### Use KUTTL for:
- ✅ Testing Helm chart deployment
- ✅ End-to-end testing with full cluster
- ✅ Testing upgrade/downgrade scenarios
- ✅ Testing multi-component interactions
- ✅ Testing in CI/CD with real cluster (kind)

## Next Steps

1. ✅ **Install KUTTL** - Follow installation instructions above
2. ✅ **Create Test Structure** - Tests have been scaffolded in `tests/` directory
3. ⬜ **Implement Tests** - Complete the test scenarios
4. ⬜ **Add Makefile Targets** - Add `make test-kuttl` target
5. ⬜ **Integrate CI/CD** - Add KUTTL tests to GitHub Actions
6. ⬜ **Add Metrics/API** - Expose controller state for verification (optional)

## References

- [KUTTL GitHub Repository](https://github.com/kudobuilder/kuttl)
- [KUTTL Documentation](https://kuttl.dev/)
- [KUTTL Getting Started](https://kuttl.dev/docs/)
- [KUTTL Test Structure](https://kuttl.dev/docs/test-structure/)
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - ⭐ Quick decision guide
- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL FAQ](./KUTTL_FAQ.md) - Frequently asked questions