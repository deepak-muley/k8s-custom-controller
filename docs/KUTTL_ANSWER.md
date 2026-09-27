# KUTTL: Where to Use It and Helm Chart Testing

## Quick Answer

### ✅ YES, You Can Use KUTTL for Helm Chart Testing!

**KUTTL is excellent for Helm chart testing** - it's actually one of the primary use cases for KUTTL!

## Where Can You Use KUTTL?

### 1. ✅ Helm Chart Testing (Primary Use Case)

**What KUTTL can test for your Helm chart:**

1. **Deployment Verification**
   - Verify Helm chart deploys correctly
   - Check that all resources are created (Deployment, ServiceAccount, RBAC, ConfigMap, NetworkPolicy)
   - Verify resources are in correct state (ready, configured correctly)

2. **Configuration Testing**
   - Test with different `values.yaml` configurations
   - Test with different namespaces
   - Test with different controller settings (namespace, GVKs)

3. **Upgrade/Downgrade Testing**
   - Test upgrading Helm chart from v1.0.0 to v1.1.0
   - Test downgrading from v1.1.0 to v1.0.0
   - Verify backward compatibility
   - Test migration scenarios

4. **Uninstall Testing**
   - Verify Helm chart uninstalls cleanly
   - Ensure all resources are removed correctly

**Example:**
```bash
# Install your Helm chart
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --set controller.namespace=test \
  --set controller.gvks="apps/v1/Deployment"

# Run KUTTL tests to verify resources
kubectl kuttl test tests/helm-deployment/

# KUTTL will verify:
# - Deployment exists and has readyReplicas: 1
# - ServiceAccount exists
# - ClusterRole and ClusterRoleBinding exist
# - ConfigMap exists (if configured)
# - NetworkPolicy exists (if configured)
```

### 2. ✅ End-to-End Controller Testing

**What you can test:**

- Deploy controller via Helm → Create test resources → Verify controller counts them
- Test namespace filtering (controller watches one namespace, ignores others)
- Test resource creation → counting → deletion workflow
- Test multiple GVKs simultaneously
- Test controller restart and recovery

**Example Test Flow:**
```
1. Deploy controller via Helm chart
2. Create test deployments in watched namespace
3. Verify controller pod is running
4. Verify controller counted the deployments (via logs/metrics/API)
5. Delete deployments
6. Verify count decreased
```

### 3. ✅ Integration Testing

**What you can test:**

- Controller + other Kubernetes resources working together
- Multi-namespace scenarios
- RBAC and security policies
- Network policies (if configured)
- Complete workflows from deployment to functionality

### 4. ✅ Anything Else with Kubernetes Resources

**KUTTL can test any Kubernetes resources:**

- Custom operators and CRDs
- Multi-component systems
- Configuration changes
- Resource dependencies
- Error handling and recovery
- Rollback scenarios

## KUTTL vs Other Testing Tools

### Comparison Table

| Tool | Purpose | Speed | Cluster Needed | Best For |
|------|---------|-------|----------------|----------|
| **Ginkgo/Gomega** | Unit tests (Go code) | ⚡ Fast | ❌ No | Counter logic, parsing |
| **envtest** | Controller integration | ⚡ Fast (1-3s) | ❌ No (in-process) | Controller logic, informers |
| **KUTTL** | Helm charts, E2E | ⚠️ Slower (10-30s) | ✅ Yes (kind) | Helm deployment, E2E workflows |
| **helm-unittest** | Template validation | ⚡ Fast | ❌ No | Template rendering (static) |
| **kubesec** | Security scanning | ⚡ Fast | ❌ No | Security analysis |
| **Pluto** | API version validation | ⚡ Fast | ❌ No | Deprecated API detection |

### When to Use Each

```
Use Ginkgo/Gomega when:
✅ Testing Go code logic
✅ Fast feedback needed
✅ No cluster needed

Use envtest when:
✅ Testing controller code
✅ Testing informers and discovery
✅ Testing resource counting
✅ Fast feedback needed

Use KUTTL when:
✅ Testing Helm chart deployment
✅ Testing E2E workflows
✅ Testing upgrades/downgrades
✅ Full cluster testing needed
✅ CI/CD E2E pipeline
```

## How KUTTL Works for Helm Chart Testing

### Test Structure

```
tests/
├── helm-deployment/              # Test Helm chart deployment
│   ├── 00-install.yaml          # Install Helm chart (via script or manual)
│   └── 01-assert-resources.yaml # Verify all resources exist
│
└── kuttl-test.yaml              # KUTTL configuration
```

### Example Test Flow

```yaml
# Step 1: Install Helm chart (before running KUTTL)
# helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
#   --namespace kuttl-test-helm

# Step 2: KUTTL verifies resources exist and are correct
# tests/helm-deployment/01-assert-resources.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1    # KUTTL waits for this
  replicas: 1

apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
  # KUTTL verifies this exists

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-custom-controller
  # KUTTL verifies RBAC exists

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-custom-controller
  # KUTTL verifies binding exists
```

## What You Can Test with KUTTL

### ✅ For Your Helm Chart

1. **All Resources Created**
   - ✅ Deployment (with correct replicas, image, security contexts)
   - ✅ ServiceAccount (with correct name)
   - ✅ ClusterRole (with correct permissions)
   - ✅ ClusterRoleBinding (with correct binding)
   - ✅ ConfigMap (if configured)
   - ✅ NetworkPolicy (if configured)
   - ✅ Service (if configured)

2. **Resource State**
   - ✅ Deployment has readyReplicas = replicas
   - ✅ Pods are Running
   - ✅ RBAC is correct
   - ✅ Security contexts are applied

3. **Configuration Variations**
   - ✅ Test with different namespaces
   - ✅ Test with different GVKs
   - ✅ Test with different replica counts
   - ✅ Test with different images

4. **End-to-End Workflows**
   - ✅ Deploy chart → Verify resources → Create test resources → Verify controller works
   - ✅ Upgrade from v1.0.0 to v1.1.0 → Verify functionality
   - ✅ Uninstall → Verify cleanup

## Running KUTTL Tests

### Installation

```bash
# Install KUTTL
make kuttl-install

# Or manually
curl -L https://github.com/kudobuilder/kuttl/releases/latest/download/kubectl-kuttl_$(uname -s)_$(uname -m) -o kubectl-kuttl
chmod +x kubectl-kuttl
sudo mv kubectl-kuttl /usr/local/bin/kubectl-kuttl
```

### Running Tests

```bash
# Run all KUTTL tests
make kuttl-test

# Run with kind cluster (auto-creates/destroys)
make kuttl-test-kind

# Run only Helm chart tests
make kuttl-test-helm

# Run specific test suite
kubectl kuttl test tests/helm-deployment/

# Cleanup test namespaces
make kuttl-clean
```

## Complete Example: KUTTL Test for Your Helm Chart

### Test Directory Structure

```
tests/
├── helm-deployment/
│   ├── 00-install.yaml       # Placeholder (install via script)
│   └── 01-assert-resources.yaml
│       ├── Deployment        # Verify Deployment exists and is ready
│       ├── ServiceAccount    # Verify SA exists
│       ├── ClusterRole       # Verify CR exists
│       └── ClusterRoleBinding # Verify CRB exists
```

### Actual Test File

```yaml
# tests/helm-deployment/01-assert-resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
spec:
  replicas: 1
status:
  readyReplicas: 1
  replicas: 1
  conditions:
  - type: Available
    status: "True"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-custom-controller
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["list", "watch", "get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-custom-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-custom-controller
subjects:
- kind: ServiceAccount
  name: k8s-custom-controller
  namespace: kuttl-test-helm
```

### Running the Test

```bash
# 1. Install Helm chart
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --create-namespace \
  --set controller.namespace=kuttl-test-controller \
  --set controller.gvks="apps/v1/Deployment"

# 2. Run KUTTL test
kubectl kuttl test tests/helm-deployment/ --start-kind=false

# 3. KUTTL will:
#    - Wait for Deployment to have readyReplicas: 1
#    - Verify ServiceAccount exists
#    - Verify ClusterRole exists
#    - Verify ClusterRoleBinding exists
#    - Report success or failure
```

## Benefits of Using KUTTL for Helm Chart Testing

### ✅ Advantages

1. **Declarative Testing** - Write tests as YAML, easy to understand
2. **Real Cluster Testing** - Tests actual Helm deployment in real cluster
3. **Wait for State** - Automatically waits for resources to be ready
4. **CI/CD Friendly** - Easy to integrate with GitHub Actions
5. **Comprehensive** - Test complete deployment workflows
6. **Upgrade Testing** - Test Helm chart upgrades between versions

### ⚠️ Considerations

1. **Slower** - Takes 10-30 seconds (needs cluster)
2. **Requires Cluster** - Needs kind or existing cluster
3. **More Complex** - More setup than unit tests
4. **Limited to Resources** - Harder to verify internal controller state

## Recommendation for Your Project

### Current Testing Status

- ✅ **Unit Tests (Ginkgo)** - DONE - Fast, comprehensive
- ✅ **Integration Tests (envtest)** - DONE - Real API server, fast
- ✅ **Static Analysis (kubesec, Pluto)** - DONE - Security and API validation
- ⬜ **E2E Tests (KUTTL)** - STRUCTURE CREATED - Optional but recommended

### Should You Add KUTTL?

**Add KUTTL if you want:**
- ✅ Helm chart E2E testing in CI/CD
- ✅ Upgrade/downgrade testing
- ✅ Complete workflow testing
- ✅ Full cluster testing before release

**You don't need KUTTL if:**
- ✅ envtest covers your needs (it does for controller logic)
- ✅ You want fast feedback (unit/integration tests are faster)
- ✅ You don't need full cluster testing

### Recommendation

**Add KUTTL for Helm chart testing** because:
1. Your Helm chart deployment is important - verify it works
2. Upgrade testing is valuable - test version migrations
3. CI/CD E2E pipeline adds confidence - catch issues before release
4. Tests serve as documentation - examples of how to use your chart

**Priority: Medium** - Current tests are sufficient, but KUTTL adds value for Helm chart testing.

## Summary

### ✅ YES, Use KUTTL for:

1. **Helm Chart Testing** ⭐ **Primary Use Case**
   - Deploy chart → Verify resources → Test functionality

2. **End-to-End Testing**
   - Complete workflows from deployment to functionality

3. **Upgrade/Downgrade Testing**
   - Test Helm chart version migrations

4. **Integration Testing**
   - Multi-component, multi-namespace scenarios

### ✅ KUTTL Test Structure Created

- `tests/helm-deployment/` - Helm chart deployment tests
- `tests/controller-functionality/` - Controller E2E tests
- `tests/namespace-filtering/` - Namespace filtering tests
- `tests/kuttl-test.yaml` - KUTTL configuration

### ✅ Makefile Targets Added

- `make kuttl-install` - Install KUTTL
- `make kuttl-test` - Run KUTTL tests
- `make kuttl-test-kind` - Run with kind cluster
- `make kuttl-test-helm` - Run Helm chart tests only
- `make kuttl-clean` - Cleanup test namespaces

## References

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl) - Official repository
- [KUTTL Documentation](https://kuttl.dev/) - Official documentation
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - ⭐ Quick decision guide
- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL Usage](./KUTTL_USAGE.md) - Detailed usage guide
- [KUTTL Guide](./KUTTL_GUIDE.md) - Complete guide
- [KUTTL Summary](./KUTTL_SUMMARY.md) - Quick reference
- [KUTTL FAQ](./KUTTL_FAQ.md) - Frequently asked questions
- [Where to Use KUTTL](./WHERE_TO_USE_KUTTL.md) - When and where to use
