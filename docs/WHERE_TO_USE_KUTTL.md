# Where to Use KUTTL - Quick Answer

## Direct Answer to Your Question

### ✅ YES, you can use KUTTL for Helm Chart Testing!

**KUTTL is excellent for Helm chart testing.** In fact, Helm chart testing is one of the primary use cases for KUTTL.

## Where Can You Use KUTTL?

### 1. ✅ Helm Chart Testing (Primary Use Case)

**What you can test:**

1. **Helm Chart Deployment**
   - Install your Helm chart and verify all resources are created
   - Test that Deployment, ServiceAccount, RBAC, ConfigMap are created correctly
   - Verify security contexts and RBAC permissions are applied

2. **Helm Chart Configuration Testing**
   - Test with different `values.yaml` configurations
   - Test with different namespaces
   - Test with different controller configurations

3. **Helm Chart Upgrade/Downgrade Testing**
   - Test upgrading from v1.0.0 to v1.1.0
   - Test downgrading from v1.1.0 to v1.0.0
   - Verify backward compatibility
   - Test migration scenarios

4. **Helm Chart Uninstall Testing**
   - Verify clean uninstallation
   - Ensure all resources are removed correctly

**Example:**
```bash
# 1. Install Helm chart
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --set controller.namespace=test \
  --set controller.gvks="apps/v1/Deployment"

# 2. Run KUTTL tests to verify resources exist and are correct
kubectl kuttl test tests/helm-deployment/

# 3. KUTTL will verify:
#    - Deployment exists and is ready
#    - ServiceAccount exists
#    - RBAC (ClusterRole, ClusterRoleBinding) exists
#    - ConfigMap exists (if configured)
```

### 2. ✅ End-to-End Controller Testing

**What you can test:**

1. **Complete Workflows**
   - Deploy controller via Helm → Create test resources → Verify counting
   - Test namespace filtering (watch one namespace, ignore others)
   - Test resource creation → counting → deletion workflow

2. **Multi-Component Testing**
   - Test controller with other Kubernetes resources
   - Test cross-namespace functionality
   - Test RBAC and security policies

3. **Controller Lifecycle**
   - Test controller restart and recovery
   - Test controller upgrade scenarios
   - Test configuration changes

**Example:**
```bash
# Deploy controller, create deployments, verify counting
kubectl kuttl test tests/controller-functionality/
```

### 3. ✅ Integration Testing

**What you can test:**

- Controller + Resources working together
- Multi-namespace scenarios
- RBAC and security policies
- Network policies (if configured)
- Complete workflows from deployment to functionality

## Can You Use KUTTL for Helm Chart Testing?

### ✅ YES! Here's How:

#### Method 1: Test Rendered Helm Templates

```bash
# 1. Render Helm chart to YAML
helm template test-controller ./helm/k8s-custom-controller/ \
  --namespace kuttl-test-helm \
  --set controller.namespace=test \
  > tests/helm-deployment/00-installed-resources.yaml

# 2. Create KUTTL test to verify resources
# tests/helm-deployment/01-assert-resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1
  # KUTTL will wait for this state
```

#### Method 2: Test Actual Helm Installation (Recommended)

```bash
# 1. Install Helm chart before tests
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --create-namespace

# 2. Run KUTTL tests to verify resources
kubectl kuttl test tests/helm-deployment/ --start-kind=false

# 3. Cleanup
helm uninstall test-controller --namespace kuttl-test-helm
```

#### Method 3: Use KUTTL Test Scripts

```yaml
# tests/helm-deployment/00-install.sh (if KUTTL supports scripts)
#!/bin/bash
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --create-namespace \
  --set controller.namespace=kuttl-test-controller \
  --set controller.gvks="apps/v1/Deployment"
```

## KUTTL vs Other Tools for Helm Testing

| Tool | Purpose | When to Use |
|------|---------|-------------|
| **KUTTL** | E2E testing of deployed Helm charts | ✅ Test complete deployment |
| **helm-unittest** | Template validation (static) | Test template rendering |
| **helm lint** | Syntax validation | Check template syntax |
| **kubesec** | Security scanning | Scan rendered templates |
| **Pluto** | API version validation | Check deprecated APIs |
| **envtest** | Controller code testing | Test controller logic |

## What KUTTL Can Test for Your Helm Chart

### ✅ Resources Created by Helm Chart

1. **Deployment**
   ```yaml
   # tests/helm-deployment/01-assert-resources.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: k8s-custom-controller
     namespace: kuttl-test-helm
   status:
     readyReplicas: 1
     replicas: 1
   ```

2. **ServiceAccount**
   ```yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: k8s-custom-controller
     namespace: kuttl-test-helm
   ```

3. **RBAC (ClusterRole, ClusterRoleBinding)**
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: k8s-custom-controller
   rules:
   - apiGroups: ["*"]
     resources: ["*"]
     verbs: ["list", "watch", "get"]
   ```

4. **ConfigMap** (if configured)
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: k8s-custom-controller-config
     namespace: kuttl-test-helm
   ```

5. **NetworkPolicy** (if configured)
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: k8s-custom-controller
     namespace: kuttl-test-helm
   ```

### ✅ Configuration Testing

- Test with different `values.yaml` configurations
- Test with different namespaces
- Test with different controller settings
- Test security contexts and RBAC

## Example: Complete KUTTL Test for Your Helm Chart

### Test Scenario: Helm Chart Deployment

```bash
# Directory structure:
tests/
└── helm-deployment/
    ├── 00-install.yaml       # Install Helm chart (via script or manual)
    ├── 01-assert-resources.yaml  # Verify all resources exist
    └── 02-test-functionality.yaml # Test controller works after deployment
```

### Step 1: Install Helm Chart

```bash
# Manual installation before KUTTL (or use script)
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --create-namespace \
  --set controller.namespace=kuttl-test-controller \
  --set controller.gvks="apps/v1/Deployment"
```

### Step 2: Verify Resources (KUTTL Test)

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
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-custom-controller
```

### Step 3: Test Controller Functionality

```yaml
# tests/helm-deployment/steps/00-create-test-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: kuttl-test-controller
spec:
  replicas: 1
  # ... deployment spec
```

```yaml
# tests/helm-deployment/steps/01-verify-controller-counts.yaml
# Verify controller counted the deployment
# (Requires querying controller state via metrics/API/logs)
```

## KUTTL Can Also Test:

### ✅ Anything Else with Kubernetes Resources

1. **Operators** - Test custom operators and their behavior
2. **CRDs** - Test custom resource definitions and controllers
3. **Multi-Component Systems** - Test complex deployments with multiple components
4. **Upgrade Scenarios** - Test upgrading between versions
5. **Rollback Scenarios** - Test rolling back to previous versions
6. **Configuration Changes** - Test updating configurations
7. **Error Handling** - Test error scenarios and recovery
8. **Resource Dependencies** - Test resource creation order and dependencies

## Summary

### ✅ Use KUTTL for:

1. **Helm Chart E2E Testing** ⭐ **Primary Use Case**
   - Deploy Helm chart → Verify resources → Test functionality

2. **End-to-End Workflows**
   - Complete deployment workflows
   - Multi-component interactions

3. **Upgrade/Downgrade Testing**
   - Helm chart version upgrades
   - Backward compatibility

4. **Integration Testing**
   - Controller with other resources
   - Multi-namespace scenarios

### ✅ Your Current Testing Strategy

```
Unit Tests (Ginkgo) ✅          → Fast, comprehensive
Integration Tests (envtest) ✅   → Real API server
E2E Tests (KUTTL) ⬜             → Full cluster (optional)
Static Analysis ✅               → Security, API versions
```

**Recommendation:** Add KUTTL tests if you want:
- Helm chart E2E testing in CI/CD
- Upgrade/downgrade testing
- Complete workflow testing
- Full cluster testing with kind

## Next Steps

1. ✅ **KUTTL test structure created** - `tests/` directory with examples
2. ✅ **Makefile targets added** - `make kuttl-install`, `make kuttl-test`, etc.
3. ✅ **Documentation created** - Complete guides and examples
4. ⬜ **Implement tests** - Complete test scenarios based on your needs
5. ⬜ **Add to CI/CD** - Optional: Add to GitHub Actions

## References

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl)
- [KUTTL Documentation](https://kuttl.dev/)
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - ⭐ Quick decision guide
- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL Usage](./KUTTL_USAGE.md) - Detailed usage
- [KUTTL Guide](./KUTTL_GUIDE.md) - Complete guide
- [KUTTL Summary](./KUTTL_SUMMARY.md) - Quick reference
- [KUTTL FAQ](./KUTTL_FAQ.md) - Frequently asked questions
