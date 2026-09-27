# KUTTL (Kubernetes Test TooL) Guide

## What is KUTTL?

[KUTTL](https://github.com/kudobuilder/kuttl) (KUbernetes Test TooL) is a **declarative testing framework** for Kubernetes that allows you to write tests using YAML manifests and test assertions. It's designed for testing Kubernetes operators and Helm charts by deploying them to a cluster and verifying the resulting state.

Reference: [KUTTL GitHub Repository](https://github.com/kudobuilder/kuttl)

## Where Can You Use KUTTL?

### 1. ✅ Helm Chart Testing (Primary Use Case)

**KUTTL is excellent for testing Helm charts** because it can:
- Deploy your Helm chart to a test cluster
- Verify that all resources are created correctly
- Test end-to-end functionality of deployed resources
- Verify resource relationships and dependencies
- Test upgrade/downgrade scenarios
- Test with different values.yaml configurations

**How it differs from other Helm testing tools:**
- **helm-unittest**: Tests Helm templates statically (template rendering validation)
- **KUTTL**: Tests Helm charts dynamically (actual deployment and behavior)

### 2. ✅ Operator Testing

- Test custom resource creation and reconciliation
- Verify controller behavior with actual Kubernetes API
- Test operator lifecycle (install, upgrade, uninstall)
- Test error handling and edge cases

### 3. ✅ End-to-End (E2E) Testing

- Test complete application deployments
- Verify resource interactions and dependencies
- Test multi-component systems
- Integration testing across namespaces

### 4. ✅ Resource Validation Testing

- Verify resources are created with correct configurations
- Test RBAC and security policies
- Verify network policies work correctly
- Test ConfigMaps and Secrets

## KUTTL vs Other Testing Approaches

| Tool | Purpose | When to Use |
|------|---------|-------------|
| **KUTTL** | Declarative E2E testing | Helm charts, operators, full deployments |
| **envtest** | Controller logic testing | Go controller code, informers, business logic |
| **helm-unittest** | Template validation | Helm template rendering, static validation |
| **kind/testcontainers** | Full cluster testing | When you need kubelet, pod scheduling |

## When to Use KUTTL for Your Project

### ✅ Use KUTTL for:

1. **Helm Chart Integration Testing**
   - Test that your Helm chart deploys correctly
   - Verify all resources are created (Deployment, ServiceAccount, RBAC, etc.)
   - Test with different values.yaml configurations
   - Verify controller can connect to cluster after deployment

2. **End-to-End Controller Testing**
   - Deploy controller via Helm chart
   - Create test resources (Deployments, Services)
   - Verify controller counts resources correctly
   - Test namespace filtering functionality

3. **Upgrade/Downgrade Testing**
   - Test Helm chart upgrades between versions
   - Verify backward compatibility
   - Test migration scenarios

4. **Multi-Component Testing**
   - Test controller + other Kubernetes resources together
   - Verify resource dependencies
   - Test complete workflows

### ❌ Don't Use KUTTL for:

1. **Unit Testing Controller Logic** - Use envtest (already implemented)
2. **Template Syntax Validation** - Use helm-unittest or helm lint
3. **Static Security Scanning** - Use kubesec (already integrated)
4. **API Version Checking** - Use Pluto (already integrated)

## How KUTTL Works

```
┌─────────────────────────────────────────────────────────┐
│                    Test Structure                        │
│  tests/                                                  │
│    ├── <test-name>/                                      │
│    │   ├── 00-install.yaml      # Setup resources       │
│    │   ├── 01-assert.yaml       # Verify state          │
│    │   ├── 02-test.yaml         # Test scenario         │
│    │   └── steps/                # Ordered test steps    │
│    │       ├── 00-step.yaml                              │
│    │       └── 01-assert.yaml                            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  KUTTL Test Runner                       │
│  1. Create test namespace                                │
│  2. Apply 00-install.yaml (setup)                       │
│  3. Wait for assertions (01-assert.yaml)                │
│  4. Apply test scenarios (02-test.yaml)                 │
│  5. Verify final state                                   │
│  6. Cleanup test namespace                               │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  Kubernetes Cluster                      │
│  (kind, minikube, or existing cluster)                  │
│  • Resources deployed                                    │
│  • Controller running                                    │
│  • Test resources created                                │
│  • State verified                                        │
└─────────────────────────────────────────────────────────┘
```

## KUTTL Test Structure

### Basic Test Directory Structure

```
tests/
├── helm-deployment/           # Test Helm chart deployment
│   ├── 00-install.yaml       # Install Helm chart
│   ├── 01-assert.yaml        # Assert resources exist
│   └── 02-cleanup.yaml       # Cleanup (optional)
│
├── controller-functionality/  # Test controller behavior
│   ├── 00-setup.yaml         # Setup test namespace
│   ├── 01-deploy-controller.yaml  # Deploy controller
│   ├── 02-create-resources.yaml   # Create test resources
│   ├── 03-assert-count.yaml       # Verify counting
│   └── steps/                     # Ordered steps
│       ├── 00-create-deployment.yaml
│       ├── 01-assert-count.yaml
│       ├── 02-delete-deployment.yaml
│       └── 03-assert-count.yaml
│
└── namespace-filtering/       # Test namespace filtering
    ├── 00-setup.yaml
    ├── 01-deploy-controller.yaml
    └── steps/
        ├── 00-create-namespace-a.yaml
        ├── 01-create-namespace-b.yaml
        ├── 02-create-resources.yaml
        └── 03-assert-filtering.yaml
```

### KUTTL Test File Format

```yaml
# tests/controller-functionality/steps/00-create-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:latest
```

```yaml
# tests/controller-functionality/steps/01-assert-count.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-assertion
  namespace: test-controller
spec:
  containers:
  - name: kubectl
    image: bitnami/kubectl:latest
    command:
    - /bin/sh
    - -c
    - |
      # Verify controller counted the deployment
      # This is a simplified example - actual assertion would use KUTTL's assertion features
      kubectl get deployments -n test-controller
---
# KUTTL assertions use special annotations
apiVersion: v1
kind: ConfigMap
metadata:
  name: assertion-check
  namespace: test-controller
  annotations:
    kuttl.assert/that.apps_v1_Deployment.test-deployment.is.present: "true"
```

## Installing KUTTL

```bash
# Install KUTTL
curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_$(uname -s)_$(uname -m) -o kubectl-kuttl
chmod +x kubectl-kuttl
sudo mv kubectl-kuttl /usr/local/bin/kubectl-kuttl

# Or using kubectl plugin
kubectl krew install kuttl

# Or via Homebrew (macOS)
brew install kuttl
```

## Running KUTTL Tests

```bash
# Run all tests in tests/ directory
kubectl kuttl test tests/

# Run specific test
kubectl kuttl test tests/helm-deployment

# Run with verbose output
kubectl kuttl test tests/ --debug

# Run against specific cluster (uses current kubeconfig context)
kubectl kuttl test tests/ --start-kind=false

# Run with cleanup
kubectl kuttl test tests/ --delete-namespace-on-success
```

## Integration with Your Project

### Current Testing Strategy

```
┌─────────────────────────────────────────────────────────┐
│                  Unit Tests (Ginkgo)                    │
│  • Counter logic                                        │
│  • Config validation                                    │
│  • GVK parsing                                          │
│  ✅ Already implemented                                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│            Integration Tests (envtest)                  │
│  • Controller lifecycle                                 │
│  • Informer cache syncing                               │
│  • Discovery API                                        │
│  • Resource counting                                    │
│  ✅ Already implemented                                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              E2E Tests (KUTTL) ⭐                       │
│  • Helm chart deployment                                │
│  • Full controller deployment                           │
│  • End-to-end workflows                                 │
│  • Multi-component testing                              │
│  ⬜ Recommended to add                                 │
└─────────────────────────────────────────────────────────┘
```

## Example: KUTTL Test for Your Helm Chart

Here's how you would structure KUTTL tests for your project:

### Test 1: Helm Chart Deployment

```yaml
# tests/helm-deployment/00-install.yaml
# This would be handled by KUTTL test harness
# Actual Helm install would be done via script or step
```

### Test 2: Controller Functionality

```yaml
# tests/controller-counting/steps/00-create-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment-1
  namespace: test-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:latest
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment-2
  namespace: test-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:latest
```

```yaml
# tests/controller-counting/steps/01-assert-count.yaml
# KUTTL will assert that resources exist and are in expected state
# Actual counting verification would need to be done via:
# 1. Query controller metrics/endpoints
# 2. Check controller logs
# 3. Use a test pod to query controller API (if exposed)
```

## When to Add KUTTL Tests

**Add KUTTL tests when you need:**

1. **Full Helm Chart Testing** - Verify complete deployment workflow
2. **End-to-End Scenarios** - Test complete user workflows
3. **Integration Testing** - Test controller with other components
4. **Upgrade Testing** - Test Helm chart version upgrades
5. **Multi-Namespace Testing** - Test cross-namespace functionality

**Current recommendation:** You already have excellent unit tests (Ginkgo) and integration tests (envtest). Add KUTTL tests if you want to:
- Test the complete Helm chart deployment in a real cluster
- Test end-to-end workflows that span multiple components
- Test upgrade/downgrade scenarios
- Test in CI/CD with actual cluster (kind)

## Benefits of Adding KUTTL

1. **Helm Chart Validation** - Ensure your Helm chart works in real clusters
2. **E2E Confidence** - Test complete workflows from deployment to functionality
3. **CI/CD Integration** - Run E2E tests in CI with kind or test clusters
4. **Regression Testing** - Catch integration issues before release
5. **Documentation** - Tests serve as examples of how to use your chart

## Next Steps

Would you like me to:
1. Create example KUTTL tests for your Helm chart?
2. Set up Makefile targets for running KUTTL tests?
3. Integrate KUTTL tests into your CI/CD pipeline?
4. Create a complete test suite structure?
