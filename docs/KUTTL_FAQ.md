# KUTTL FAQ - Direct Answers

## Q: Can I use KUTTL for Helm chart testing?

### ✅ YES! KUTTL is excellent for Helm chart testing!

**Helm chart testing is one of the primary use cases for KUTTL.** Here's why:

1. **Deploy Helm Chart** - Install your Helm chart to a test cluster
2. **Verify Resources** - KUTTL automatically verifies all resources (Deployment, ServiceAccount, RBAC, etc.) exist and are in correct state
3. **Test Functionality** - Verify controller works after deployment
4. **Test Upgrades** - Test Helm chart upgrades between versions
5. **Test Configurations** - Test with different values.yaml configurations

**Example:**
```bash
# 1. Install Helm chart
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
  --namespace kuttl-test-helm \
  --set controller.namespace=test \
  --set controller.gvks="apps/v1/Deployment"

# 2. Run KUTTL test to verify resources
kubectl kuttl test tests/helm-deployment/

# 3. KUTTL will verify:
#    ✅ Deployment exists and readyReplicas: 1
#    ✅ ServiceAccount exists
#    ✅ ClusterRole exists
#    ✅ ClusterRoleBinding exists
#    ✅ ConfigMap exists (if configured)
```

## Q: Where can I use KUTTL?

### ✅ Use Cases for KUTTL

1. **Helm Chart Testing** ⭐ **Primary Use Case**
   - Test complete Helm chart deployment
   - Verify all resources are created correctly
   - Test upgrade/downgrade scenarios
   - Test with different configurations

2. **End-to-End Testing**
   - Test complete workflows from deployment to functionality
   - Test multi-component interactions
   - Test cross-namespace functionality

3. **Operator Testing**
   - Test Kubernetes operators and their behavior
   - Test CRD controllers
   - Test reconciliation logic

4. **Integration Testing**
   - Test multiple components working together
   - Test resource dependencies
   - Test configuration changes

5. **Anything with Kubernetes Resources**
   - Test any Kubernetes resources declaratively
   - Test resource creation, updates, deletions
   - Test resource state and relationships

## Q: Can I use KUTTL for anything else besides Helm charts?

### ✅ YES! KUTTL can test:

1. **Any Kubernetes Resources**
   - Deployments, Services, ConfigMaps, Secrets
   - Custom resources (CRDs)
   - RBAC (ClusterRole, ClusterRoleBinding, Role, RoleBinding)
   - NetworkPolicies, ServiceAccounts
   - Namespaces, PersistentVolumes, etc.

2. **Complete Applications**
   - Multi-component deployments
   - Resource dependencies
   - Configuration management

3. **Operators**
   - Custom operators
   - CRD controllers
   - Reconciliation logic

4. **Upgrade/Downgrade Scenarios**
   - Helm chart upgrades
   - Operator upgrades
   - Configuration migrations

## Q: How does KUTTL compare to envtest?

### Comparison: KUTTL vs envtest

| Aspect | envtest | KUTTL |
|--------|---------|-------|
| **Purpose** | Controller code testing | Helm charts, E2E workflows |
| **What it tests** | Go code, informers, discovery | Helm deployment, complete workflows |
| **Cluster needed** | ❌ No (in-process API server) | ✅ Yes (kind or existing) |
| **Speed** | ⚡ Fast (1-3 seconds) | ⚠️ Slower (10-30 seconds) |
| **Helm chart testing** | ❌ No | ✅ Yes |
| **Upgrade testing** | ❌ No | ✅ Yes |
| **Declarative** | ❌ Code-based | ✅ YAML-based |
| **CI/CD friendly** | ✅ Yes (no Docker) | ⚠️ Yes (but needs cluster) |

### When to Use Each

**Use envtest for:**
- ✅ Controller logic testing (already implemented)
- ✅ Informer cache syncing
- ✅ Discovery API interactions
- ✅ Fast feedback during development

**Use KUTTL for:**
- ✅ Helm chart deployment testing
- ✅ Complete E2E workflows
- ✅ Upgrade/downgrade testing
- ✅ Full cluster testing

## Q: What's the difference between KUTTL and helm-unittest?

### KUTTL vs helm-unittest

| Aspect | helm-unittest | KUTTL |
|--------|---------------|-------|
| **Purpose** | Template validation | E2E deployment testing |
| **What it tests** | Template rendering (static) | Actual deployment (dynamic) |
| **Cluster needed** | ❌ No | ✅ Yes |
| **When** | Before deployment | After deployment |
| **Speed** | ⚡ Fast (template rendering) | ⚠️ Slower (cluster operations) |
| **Tests** | Template output | Actual Kubernetes resources |

**Example:**

```bash
# helm-unittest: Test template rendering (static)
helm unittest ./helm/k8s-custom-controller/
# ✅ Tests that templates render correctly
# ✅ Tests that values are substituted correctly
# ❌ Doesn't test actual deployment

# KUTTL: Test actual deployment (dynamic)
helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz
kubectl kuttl test tests/helm-deployment/
# ✅ Tests that resources are created in cluster
# ✅ Tests that resources are in correct state
# ✅ Tests that controller works after deployment
```

**Recommendation:** Use both!
- **helm-unittest** - Fast template validation during development
- **KUTTL** - E2E deployment testing in CI/CD

## Q: Do I need KUTTL if I already have envtest?

### Answer: It Depends on What You Want to Test

**You DON'T need KUTTL if:**
- ✅ envtest covers your needs (controller logic) - **You have this!**
- ✅ You want fast feedback - **envtest is faster**
- ✅ You don't need Helm chart E2E testing

**You SHOULD add KUTTL if:**
- ✅ You want Helm chart E2E testing
- ✅ You want upgrade/downgrade testing
- ✅ You want full cluster testing in CI/CD
- ✅ You want complete workflow testing

### Recommendation for Your Project

```
Current Testing Pyramid:
┌─────────────────────────────────────┐
│   Unit Tests (Ginkgo) ✅            │  ← Fast, comprehensive
│   - Counter, config, parsing       │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│   Integration (envtest) ✅           │  ← Real API server
│   - Controller, informers, discovery│
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│   E2E Tests (KUTTL) ⬜                │  ← Full cluster (optional)
│   - Helm chart, upgrades, E2E      │
└─────────────────────────────────────┘

Recommendation: Add KUTTL for Helm chart E2E testing
```

## Q: How do I test Helm charts with KUTTL?

### Step-by-Step Guide

1. **Install KUTTL**
   ```bash
   make kuttl-install
   ```

2. **Create Test Structure**
   ```
   tests/
   └── helm-deployment/
       ├── 00-install.yaml       # Install Helm chart (via script)
       └── 01-assert-resources.yaml  # Verify resources exist
   ```

3. **Install Helm Chart** (before or in test)
   ```bash
   helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz \
     --namespace kuttl-test-helm \
     --create-namespace \
     --set controller.namespace=test \
     --set controller.gvks="apps/v1/Deployment"
   ```

4. **Create Assertion File**
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

5. **Run KUTTL Test**
   ```bash
   kubectl kuttl test tests/helm-deployment/
   ```

6. **KUTTL will:**
   - Wait for Deployment to have `readyReplicas: 1`
   - Verify ServiceAccount exists
   - Verify RBAC exists
   - Report success or failure

## Q: Can KUTTL test controller functionality after Helm deployment?

### ✅ YES! Here's How:

**Test Flow:**

1. **Deploy Controller via Helm**
   ```bash
   helm install test-controller ./helm/k8s-custom-controller/charts/*.tgz
   ```

2. **Wait for Controller to be Ready**
   ```yaml
   # tests/controller-functionality/00-assert-controller-ready.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: k8s-custom-controller
   status:
     readyReplicas: 1
   ```

3. **Create Test Resources**
   ```yaml
   # tests/controller-functionality/steps/00-create-deployments.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: test-deployment-1
     namespace: kuttl-test-controller
   spec:
     replicas: 1
     # ... deployment spec
   ```

4. **Verify Controller Counted Resources**
   ```yaml
   # tests/controller-functionality/steps/01-verify-count.yaml
   # This requires querying controller state via:
   # - Metrics endpoint (if added)
   # - Health endpoint (if added)
   # - Controller logs
   # - Test pod to query controller
   ```

**Note:** To verify controller counts, you need to expose controller state:
- Add `/metrics` endpoint with count metrics
- Add `/health` endpoint that returns counts
- Query controller logs for count information
- Use test pod to exec into controller and query state

## Summary: Direct Answers

### ✅ Can I use KUTTL for Helm chart testing?

**YES!** KUTTL is excellent for Helm chart testing. It's one of the primary use cases.

### ✅ Where can I use KUTTL?

1. **Helm Chart Testing** ⭐ **Primary Use Case**
   - Deploy chart → Verify resources → Test functionality
   
2. **End-to-End Testing**
   - Complete workflows from deployment to functionality

3. **Upgrade/Downgrade Testing**
   - Test Helm chart upgrades between versions

4. **Integration Testing**
   - Multi-component, multi-namespace scenarios

5. **Anything with Kubernetes Resources**
   - Test any Kubernetes resources declaratively

### ✅ Can I use KUTTL for anything else?

**YES!** KUTTL can test:
- ✅ Operators and CRDs
- ✅ Multi-component systems
- ✅ Resource dependencies
- ✅ Configuration changes
- ✅ Error handling and recovery

### ✅ What's already implemented?

- ✅ **Unit Tests (Ginkgo)** - Counter, config, parsing
- ✅ **Integration Tests (envtest)** - Controller, informers, discovery
- ✅ **Static Analysis** - Kubesec, Pluto
- ✅ **KUTTL Test Structure** - Tests scaffolded in `tests/` directory
- ✅ **Makefile Targets** - `make kuttl-install`, `make kuttl-test`, etc.
- ✅ **Documentation** - Complete guides and examples

### ✅ Next Steps

1. ✅ Test structure created
2. ✅ Makefile targets added
3. ✅ Documentation created
4. ⬜ Implement tests - Complete test scenarios (optional)
5. ⬜ Add to CI/CD - Optional: Add to GitHub Actions
6. ⬜ Add metrics/API - Expose controller state for verification (optional)

## Quick Start

```bash
# Install KUTTL
make kuttl-install

# Run tests (requires cluster or kind)
make kuttl-test

# Or run with kind cluster (auto-creates/destroys)
make kuttl-test-kind

# Run only Helm chart tests
make kuttl-test-helm

# Cleanup
make kuttl-clean
```

## References

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl)
- [KUTTL Documentation](https://kuttl.dev/)
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - ⭐ Quick decision guide
- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL Usage](./KUTTL_USAGE.md) - Detailed usage guide
- [KUTTL Guide](./KUTTL_GUIDE.md) - Complete guide
- [KUTTL Summary](./KUTTL_SUMMARY.md) - Quick reference
- [Where to Use KUTTL](./WHERE_TO_USE_KUTTL.md) - When and where to use
