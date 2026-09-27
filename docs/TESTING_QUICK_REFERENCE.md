# Testing Quick Reference Guide

## When to Use Each Type of Test

### 🎯 Quick Decision Tree

```
Do you need to test Go code logic?
├─ YES → Use Ginkgo/Gomega (Unit Tests)
│         ✅ Fast (milliseconds)
│         ✅ No cluster needed
│         ✅ Test: counter logic, config validation, parsing
│
└─ NO → Do you need to test controller with real API server?
    ├─ YES → Use envtest (Integration Tests)
    │         ✅ Fast (1-3 seconds)
    │         ✅ Real API server (in-process)
    │         ✅ Test: informers, discovery, resource counting
    │
    └─ NO → Do you need to test Helm chart deployment?
        ├─ YES → Use KUTTL (E2E Tests)
        │         ⚠️ Slower (10-30 seconds)
        │         ✅ Real cluster (kind or existing)
        │         ✅ Test: Helm deployment, upgrades, full workflows
        │
        └─ NO → Do you need static analysis?
            ├─ YES → Use kubesec + Pluto
            │         ✅ Fast (seconds)
            │         ✅ No cluster needed
            │         ✅ Test: security, API version validation
```

## Test Types Comparison

| Test Type | Tool | Speed | Cluster | When to Use | Examples |
|-----------|------|-------|---------|-------------|----------|
| **Unit Tests** | Ginkgo/Gomega | ⚡ Fast (ms) | ❌ No | Test Go code logic | Counter, config, parsing |
| **Integration Tests** | envtest | ⚡ Fast (1-3s) | ❌ No (API server in-process) | Test controller with real API | Informers, discovery, counting |
| **E2E Tests** | KUTTL | ⚠️ Slower (10-30s) | ✅ Yes (kind) | Test Helm charts, full workflows | Helm deployment, upgrades |
| **Security Scan** | kubesec | ⚡ Fast (s) | ❌ No | Scan Helm templates | Security vulnerabilities |
| **API Validation** | Pluto | ⚡ Fast (s) | ❌ No | Check API versions | Deprecated APIs |

## Detailed Decision Guide

### 1. Unit Tests (Ginkgo/Gomega) ✅ Implemented

**When to Use:**
- ✅ Testing isolated Go code logic
- ✅ Testing business logic (counter, config validation)
- ✅ Testing helper functions (parsing, formatting)
- ✅ Fast feedback during development

**When NOT to Use:**
- ❌ Testing Kubernetes API interactions
- ❌ Testing informers or controllers
- ❌ Testing resource discovery

**Example:**
```go
// ✅ GOOD: Test counter logic
Describe("ResourceCounter", func() {
    It("should increment count", func() {
        counter.Increment(gvk)
        Expect(counter.GetCount(gvk)).To(Equal(1))
    })
})

// ❌ BAD: Testing Kubernetes API (use envtest instead)
It("should create deployment", func() {
    // This needs real API server
})
```

**Files:**
- `internal/counter/counter_test.go`
- `internal/config/config_test.go`
- `main_test.go`
- `internal/controller/controller_test.go` (unit tests only)

**Run:**
```bash
make test-unit        # Run unit tests only
make test             # Run all tests (includes unit)
```

---

### 2. Integration Tests (envtest) ✅ Implemented

**When to Use:**
- ✅ Testing controller with real Kubernetes API server
- ✅ Testing informer cache syncing
- ✅ Testing discovery API interactions
- ✅ Testing resource counting (create/delete events)
- ✅ Testing namespace filtering
- ✅ Testing multiple GVKs simultaneously

**When NOT to Use:**
- ❌ Testing Helm chart deployment
- ❌ Testing pod scheduling (no kubelet)
- ❌ Testing full cluster behavior

**Example:**
```go
// ✅ GOOD: Test controller with real API server
Describe("GVKController with EnvTest", func() {
    It("should count resources correctly", func() {
        // Create deployment via dynamic client
        deployment := createDeployment(dynamicClient)
        
        // Wait for informer to sync
        Eventually(func() int {
            return controller.GetCounter().GetCount(deploymentGVK)
        }).Should(Equal(1))
    })
})
```

**Files:**
- `internal/controller/suite_test.go` - Test suite setup
- `internal/controller/controller_envtest_test.go` - Integration tests

**Run:**
```bash
make test-integration  # Run integration tests with envtest
make test             # Run all tests (includes integration)
```

**Why envtest instead of fake clients?**
- ✅ Real API server (supports dynamic client, discovery, informers)
- ✅ Real informer behavior (cache syncing, watch events)
- ✅ Real discovery API (resource pluralization, GVR resolution)
- ❌ Fake clients don't support dynamic client or discovery well

---

### 3. E2E Tests (KUTTL) ⬜ Optional

**When to Use:**
- ✅ Testing Helm chart deployment
- ✅ Testing complete deployment workflows
- ✅ Testing Helm chart upgrades/downgrades
- ✅ Testing multi-component interactions
- ✅ Testing in CI/CD with full cluster

**When NOT to Use:**
- ❌ Testing controller code logic (use envtest)
- ❌ Fast feedback (use unit/integration tests)
- ❌ Testing during development (too slow)

**Example:**
```yaml
# ✅ GOOD: Test Helm chart deployment
# tests/helm-deployment/01-assert-resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1  # KUTTL waits for this state
```

**Files:**
- `tests/helm-deployment/` - Helm chart tests
- `tests/controller-functionality/` - Controller E2E tests
- `tests/namespace-filtering/` - Namespace filtering tests

**Run:**
```bash
make kuttl-install    # Install KUTTL
make kuttl-test       # Run KUTTL tests (requires cluster)
make kuttl-test-kind  # Run with kind cluster (auto-creates)
```

**Why KUTTL instead of envtest?**
- ✅ Tests actual Helm chart deployment
- ✅ Tests complete workflows (deploy → test → cleanup)
- ✅ Tests upgrade/downgrade scenarios
- ⚠️ Slower than envtest (needs full cluster)

---

### 4. Security Scanning (kubesec) ✅ Implemented

**When to Use:**
- ✅ Before deploying Helm charts
- ✅ In CI/CD pipeline
- ✅ Validating security best practices
- ✅ Ensuring high security scores

**When NOT to Use:**
- ❌ Testing functionality (use other tests)
- ❌ Fast feedback during development

**Run:**
```bash
make kubesec-install  # Install kubesec
make kubesec-scan     # Scan Helm templates
make security-scan    # Full security scan
```

**Why kubesec?**
- ✅ Industry standard for Kubernetes security scanning
- ✅ Comprehensive security checks
- ✅ Configurable minimum score
- ✅ Easy CI/CD integration

---

### 5. API Version Validation (Pluto) ✅ Implemented

**When to Use:**
- ✅ Before Kubernetes upgrades
- ✅ In CI/CD pipeline
- ✅ Detecting deprecated APIs
- ✅ Preventing API breakage

**When NOT to Use:**
- ❌ Testing functionality (use other tests)

**Run:**
```bash
make pluto-install         # Install Pluto
make pluto-detect-validate # Validate API versions
make pluto-detect-charts   # Check Helm charts
```

**Why Pluto?**
- ✅ Detects deprecated Kubernetes API versions
- ✅ Prevents future breakage
- ✅ Supports target Kubernetes versions
- ✅ Easy integration with CI/CD

---

## Complete Testing Strategy

### Testing Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱    ╲
                 ╱ E2E  ╲        ← KUTTL (Optional)
                ╱ Tests  ╲       1. Helm chart deployment
               ╱──────────╲      2. Complete workflows
              ╱            ╲     3. Upgrade testing
             ╱ Integration  ╲    ⚠️ Slower (10-30s)
            ╱   Tests        ╲   ✅ Real cluster
           ╱──────────────────╲  Run: make kuttl-test
          ╱                    ╲
         ╱   envtest             ╲
        ╱──────────────────────────╲
       ╱                            ╲
      ╱      Unit Tests              ╲
     ╱────────────────────────────────╲
    ╱  Ginkgo/Gomega                  ╲
   ╱────────────────────────────────────╲
  ╱  Counter, Config, Parsing          ╲
 ╱────────────────────────────────────────╲
╱────────────────────────────────────────────╲
  ✅ Fast (ms)     ✅ Fast (1-3s)    ⚠️ Slower (10-30s)
  ✅ No cluster    ✅ API server     ✅ Full cluster
```

### Test Coverage

| Component | Unit Tests | Integration Tests | E2E Tests | Status |
|-----------|------------|-------------------|-----------|--------|
| Counter Logic | ✅ Ginkgo | ❌ N/A | ❌ N/A | ✅ Done |
| Config Validation | ✅ Ginkgo | ❌ N/A | ❌ N/A | ✅ Done |
| Parsing Functions | ✅ Ginkgo | ❌ N/A | ❌ N/A | ✅ Done |
| Controller Logic | ✅ Ginkgo (mocks) | ✅ envtest | ⬜ KUTTL (optional) | ✅ Done |
| Informer Syncing | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ✅ Done |
| Discovery API | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ✅ Done |
| Resource Counting | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ✅ Done |
| Helm Chart | ❌ N/A | ❌ N/A | ⬜ KUTTL | ⬜ Optional |
| Security | ❌ N/A | ❌ N/A | ✅ kubesec | ✅ Done |
| API Versions | ❌ N/A | ❌ N/A | ✅ Pluto | ✅ Done |

## Running Tests

### Quick Commands

```bash
# Unit tests only (fast, no cluster)
make test-unit

# Integration tests (fast, API server in-process)
make test-integration

# All tests (unit + integration)
make test

# E2E tests (slower, requires cluster)
make kuttl-install
make kuttl-test-kind  # Auto-creates kind cluster

# Security and API validation
make security-scan
make pluto-detect-validate

# Full CI pipeline
make ci
```

### Test Execution Time

| Test Type | Typical Time | When to Run |
|-----------|--------------|-------------|
| Unit Tests | < 1 second | Every code change |
| Integration Tests | 1-3 seconds | Before commit |
| E2E Tests (KUTTL) | 10-30 seconds | Before release, CI/CD |
| Security Scan | < 5 seconds | Before commit, CI/CD |
| API Validation | < 5 seconds | Before commit, CI/CD |

## Best Practices

### 1. Write Unit Tests First

✅ **Start with unit tests** - They're fastest and catch logic errors quickly

```go
// Write this first
Describe("ResourceCounter", func() {
    It("should increment count", func() {
        counter.Increment(gvk)
        Expect(counter.GetCount(gvk)).To(Equal(1))
    })
})
```

### 2. Add Integration Tests for API Interactions

✅ **Add integration tests** when you need real API server behavior

```go
// Add this for real API testing
Describe("GVKController with EnvTest", func() {
    It("should count resources from API server", func() {
        // Real API server test
    })
})
```

### 3. Use E2E Tests for Complete Workflows

✅ **Add E2E tests** for Helm charts and complete workflows (optional)

```yaml
# Add this for Helm chart testing
# tests/helm-deployment/01-assert-resources.yaml
```

### 4. Run Security and API Validation in CI/CD

✅ **Always run** security scans and API validation before release

```bash
# In CI/CD
make security-scan
make pluto-detect-validate
```

## Common Mistakes

### ❌ Mistake 1: Using envtest for Unit Tests

```go
// ❌ BAD: Too slow for simple logic test
Describe("Counter", func() {
    BeforeEach(func() {
        // Starting envtest for simple counter test
        testEnv = &envtest.Environment{}
        // ...
    })
    It("should increment", func() {
        // Simple logic that doesn't need API server
    })
})

// ✅ GOOD: Use unit test instead
Describe("Counter", func() {
    It("should increment", func() {
        counter.Increment(gvk)
        Expect(counter.GetCount(gvk)).To(Equal(1))
    })
})
```

### ❌ Mistake 2: Using Fake Clients for Dynamic Client Tests

```go
// ❌ BAD: Fake clients don't support dynamic client well
fakeClient := fake.NewSimpleClientset()
dynamicClient := dynamicfake.NewSimpleDynamicClient(...)
// Won't work properly for discovery or informers

// ✅ GOOD: Use envtest for real API server
testEnv := &envtest.Environment{}
cfg, _ := testEnv.Start()
dynamicClient, _ := dynamic.NewForConfig(cfg)
// Works perfectly for discovery and informers
```

### ❌ Mistake 3: Using KUTTL for Fast Feedback

```bash
# ❌ BAD: Running KUTTL for every code change
make kuttl-test  # Takes 10-30 seconds

# ✅ GOOD: Use unit/integration tests during development
make test-unit    # Takes < 1 second
make test         # Takes 1-3 seconds
```

## Summary

| Question | Answer |
|----------|--------|
| **Test Go code logic?** | → Unit Tests (Ginkgo) |
| **Test controller with API?** | → Integration Tests (envtest) |
| **Test Helm chart?** | → E2E Tests (KUTTL) |
| **Validate security?** | → kubesec |
| **Check API versions?** | → Pluto |
| **Fast feedback?** | → Unit + Integration Tests |
| **Before release?** | → All tests + Security + API validation |

## References

- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [envtest Explained](./ENVTEST_EXPLAINED.md) - Detailed envtest guide
- [KUTTL Usage](./KUTTL_USAGE.md) - KUTTL E2E testing guide
- [KUTTL FAQ](./KUTTL_FAQ.md) - KUTTL frequently asked questions
