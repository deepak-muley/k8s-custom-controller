# Testing Strategy for Kubernetes Controller

## Overview

This document explains the **comprehensive testing strategy** implemented for this Kubernetes controller, including all testing approaches chosen, why they were selected, and when to use each.

## Understanding Your Controller

Your controller uses:
- **Dynamic Client** (`dynamic.Interface`) - Works with any Kubernetes resource without compile-time type definitions
- **Discovery Client** (`discovery.DiscoveryInterface`) - Discovers available API resources at runtime
- **Dynamic Informers** (`dynamicinformer.DynamicSharedInformerFactory`) - Event-driven watchers for arbitrary resources
- **Cache Syncing** - Waits for informer caches to sync with API server state

## Complete Testing Strategy Overview

This project implements a **comprehensive hybrid testing approach** combining:

1. **Unit Tests (Ginkgo/Gomega)** ✅ - Fast, isolated Go code testing
2. **Integration Tests (envtest)** ✅ - Controller testing with real API server
3. **E2E Tests (KUTTL)** ⬜ - Helm chart and complete workflow testing (optional)
4. **Security Scanning (kubesec)** ✅ - Helm template security validation
5. **API Version Validation (Pluto)** ✅ - Deprecated API detection

## Testing Approaches Comparison

### 1. Unit Tests - Ginkgo/Gomega ✅ **CHOSEN & IMPLEMENTED**

**What it is:**
- **Ginkgo** - BDD-style testing framework for Go
- **Gomega** - Matcher library for Ginkgo
- Fast, no cluster required
- Perfect for isolated Go code logic testing

**Why we chose it:**
- ✅ **BDD-style syntax** - `Describe` and `It` blocks make tests readable
- ✅ **Rich matchers** - Gomega provides extensive assertion matchers
- ✅ **Fast execution** - Runs in milliseconds
- ✅ **No dependencies** - Doesn't require cluster or external services
- ✅ **Industry standard** - Widely used in Kubernetes projects (Kubebuilder, controller-runtime)

**What we test with it:**
- Counter logic (`internal/counter/counter_test.go`)
- Config validation (`internal/config/config_test.go`)
- Parsing functions (`main_test.go`)
- Controller unit tests with mocks (`internal/controller/controller_test.go`)

**When to use:**
- ✅ Testing isolated Go code logic
- ✅ Testing business logic (counter, config validation)
- ✅ Testing helper functions (parsing, formatting)
- ✅ Fast feedback during development

**Why NOT fake clients for unit tests?**
- We use **mocks** for controller unit tests, not fake clients
- Fake clients don't work well with dynamic clients (our controller's primary interface)
- Mocks provide better control and isolation for unit tests

**Implementation:**
- ✅ All Go files have corresponding `*_test.go` files
- ✅ Uses `Describe` and `It` blocks for BDD-style testing
- ✅ Uses Gomega matchers for assertions
- ✅ Fast execution (< 1 second for all unit tests)

**Files:**
- `internal/counter/counter_test.go`
- `internal/config/config_test.go`
- `main_test.go`
- `internal/controller/controller_test.go` (unit tests with mocks)

**Run:**
```bash
make test-unit        # Run unit tests only (fast)
make test             # Run all tests (includes unit)
```

---

### 2. Integration Tests - envtest ✅ **CHOSEN & IMPLEMENTED**

**What it is:**
- **envtest** from `sigs.k8s.io/controller-runtime/pkg/envtest`
- Provides a **real Kubernetes API server** for testing without needing a full cluster
- Includes etcd and API server binaries, but no kubelet or other node components
- Reference: [Kubebuilder Testing Guide](https://book.kubebuilder.io/cronjob-tutorial/writing-tests)

**Why we chose it:**
- ✅ **Real API server** - Supports dynamic client, discovery, and informers (critical for our controller)
- ✅ **Faster than kind** - Starts in ~1-3 seconds vs 10-30 seconds for kind
- ✅ **No Docker required** - Runs API server directly in test process
- ✅ **Perfect for dynamic clients** - Full support for `dynamic.Interface` and `discovery.DiscoveryInterface`
- ✅ **Real informers** - Actual watch events, cache syncing, and event handlers
- ✅ **CI/CD friendly** - Works in GitHub Actions without Docker
- ✅ **Kubebuilder standard** - Recommended approach for controller testing

**What we test with it:**
- Discovery API interactions
- Controller lifecycle (start/stop)
- Informer cache syncing
- Resource counting (create/delete events)
- Namespace filtering
- Multiple GVKs simultaneously

**When to use:**
- ✅ Testing controller with real Kubernetes API server
- ✅ Testing informer behavior and cache syncing
- ✅ Testing discovery API interactions
- ✅ Testing resource counting workflows
- ✅ Testing namespace filtering

**Why NOT fake clients for integration tests?**
- ❌ **No dynamic client support** - Fake clients primarily support typed clients
- ❌ **Discovery API is hard to fake** - Our controller relies heavily on `discoveryClient.ServerResourcesForGroupVersion()`
- ❌ **Informers don't work** - Dynamic informers need a real API server for proper cache syncing
- ❌ **GVR resolution fails** - Converting GVK → GVR requires real discovery API

**Why NOT kind/testcontainers for integration tests?**
- ⚠️ **Much slower** - 10-30 seconds vs 1-3 seconds for envtest
- ⚠️ **Requires Docker** - More complex setup, not ideal for CI/CD
- ⚠️ **Overkill** - We don't need kubelet or pod scheduling, only API server
- ✅ **envtest is perfect** - Provides exactly what we need (API server) without overhead

**Implementation:**
- ✅ `internal/controller/suite_test.go` - Test suite setup following Kubebuilder pattern
- ✅ `internal/controller/controller_envtest_test.go` - Comprehensive integration tests
- ✅ Tests discovery, informers, resource counting, namespace filtering

**Files:**
- `internal/controller/suite_test.go` - Test environment setup
- `internal/controller/controller_envtest_test.go` - Integration tests

**Run:**
```bash
make test-integration  # Run integration tests with envtest
make test             # Run all tests (includes integration)
```

---

### 3. E2E Tests - KUTTL ⬜ **OPTIONAL - STRUCTURE CREATED**

**What it is:**
- **KUTTL** (KUbernetes Test TooL) from `github.com/kudobuilder/kuttl`
- Declarative testing framework using YAML manifests
- Tests Helm charts and complete workflows in real clusters

**Why we considered it:**
- ✅ **Helm chart testing** - Can test complete Helm chart deployments
- ✅ **E2E workflows** - Test complete deployment → functionality → cleanup workflows
- ✅ **Upgrade testing** - Test Helm chart upgrades between versions
- ✅ **Real cluster** - Uses kind or existing cluster for realistic testing

**Why it's optional:**
- ⚠️ **Slower** - Takes 10-30 seconds (requires full cluster)
- ⚠️ **Requires cluster** - Needs kind or existing cluster
- ✅ **envtest covers controller logic** - Our integration tests already cover controller functionality
- ✅ **Static analysis covers Helm templates** - kubesec and Pluto validate Helm charts

**When to use (if implemented):**
- ✅ Testing Helm chart deployment in real cluster
- ✅ Testing upgrade/downgrade scenarios
- ✅ Testing complete workflows (deploy → test → cleanup)
- ✅ CI/CD E2E pipeline (before releases)

**What we've prepared:**
- ✅ Test structure created in `tests/` directory
- ✅ Example tests for Helm deployment, controller functionality, namespace filtering
- ✅ Makefile targets: `make kuttl-install`, `make kuttl-test`, `make kuttl-test-kind`
- ✅ Complete documentation in `docs/` folder

**Implementation status:**
- ⬜ Test structure created, ready to implement when needed
- ⬜ Not required for current testing needs (envtest covers controller logic)

**Files:**
- `tests/helm-deployment/` - Helm chart deployment tests
- `tests/controller-functionality/` - Controller E2E tests
- `tests/namespace-filtering/` - Namespace filtering tests
- `tests/kuttl-test.yaml` - KUTTL configuration

**Run (when implemented):**
```bash
make kuttl-install    # Install KUTTL
make kuttl-test       # Run KUTTL tests (requires cluster)
make kuttl-test-kind  # Run with kind cluster (auto-creates)
```

**Decision rationale:**
- ✅ **Test structure created** - Ready to use when Helm chart E2E testing is needed
- ✅ **Optional but available** - Not blocking development, can be added when required
- ✅ **Complements existing tests** - Adds value for Helm chart and upgrade testing
- ✅ **Complete documentation** - Fully documented in case it's needed

---

### 4. Security Scanning - kubesec ✅ **CHOSEN & IMPLEMENTED**

**What it is:**
- **kubesec** - Security scanner for Kubernetes manifests
- Scans Helm templates and Kubernetes resources for security vulnerabilities
- Provides security scores and recommendations

**Why we chose it:**
- ✅ **Industry standard** - Widely used for Kubernetes security scanning
- ✅ **Helm integration** - Scans rendered Helm templates
- ✅ **Comprehensive checks** - Covers security contexts, RBAC, network policies, etc.
- ✅ **Configurable minimum score** - Enforces security standards (default: 90)
- ✅ **CI/CD integration** - Easy to integrate in GitHub Actions
- ✅ **JSON reports** - Generates detailed security reports

**What we scan:**
- Rendered Helm templates
- Security contexts (runAsNonRoot, readOnlyRootFilesystem, etc.)
- RBAC configurations
- Network policies
- Resource limits and requests

**When to use:**
- ✅ Before deploying Helm charts
- ✅ In CI/CD pipeline (automated)
- ✅ Validating security best practices
- ✅ Ensuring high security scores

**Implementation:**
- ✅ Makefile target: `make kubesec-install` - Installs kubesec
- ✅ Makefile target: `make kubesec-scan` - Scans Helm templates
- ✅ Makefile target: `make security-scan` - Full security scan workflow
- ✅ CI/CD integration in GitHub Actions
- ✅ Minimum score validation (default: 90)

**Configuration:**
- Minimum score: 90 (configurable via `KUBESEC_MIN_SCORE`)
- Scans all rendered Helm templates
- Generates JSON reports for CI/CD artifacts

**Run:**
```bash
make kubesec-install  # Install kubesec
make kubesec-scan     # Scan Helm templates
make security-scan    # Full security scan (render + validate)
```

**Decision rationale:**
- ✅ **Security is critical** - Kubernetes resources must follow security best practices
- ✅ **Automated validation** - Prevents insecure configurations from being deployed
- ✅ **High minimum score** - Ensures only secure configurations pass (score ≥ 90)
- ✅ **Helm integration** - Validates actual deployed resources (rendered templates)

---

### 5. API Version Validation - Pluto ✅ **CHOSEN & IMPLEMENTED**

**What it is:**
- **Pluto** from `github.com/FairwindsOps/pluto`
- Detects deprecated Kubernetes API versions in manifests and Helm charts
- Prevents API breakage when upgrading Kubernetes clusters

**Why we chose it:**
- ✅ **Proactive detection** - Finds deprecated APIs before cluster upgrade
- ✅ **Helm chart support** - Scans Helm charts and templates
- ✅ **Target version support** - Can target specific Kubernetes versions (default: 1.29)
- ✅ **CI/CD integration** - Easy to integrate in GitHub Actions
- ✅ **Prevents breakage** - Catches API issues early before deployment

**What we validate:**
- Deprecated API versions in Helm templates
- Removed API versions that would break in target Kubernetes version
- API version compatibility

**When to use:**
- ✅ Before Kubernetes cluster upgrades
- ✅ In CI/CD pipeline (automated)
- ✅ When updating Helm charts
- ✅ Validating API version compatibility

**Implementation:**
- ✅ Makefile target: `make pluto-install` - Installs Pluto
- ✅ Makefile target: `make pluto-detect-validate` - Validates API versions
- ✅ Makefile target: `make pluto-detect-charts` - Scans Helm charts
- ✅ CI/CD integration in GitHub Actions
- ✅ Target Kubernetes version: 1.29 (configurable via `PLUTO_TARGET_K8S_VERSION`)

**Configuration:**
- Target Kubernetes version: 1.29 (configurable)
- Validates all Helm templates
- Fails build if deprecated APIs found

**Run:**
```bash
make pluto-install         # Install Pluto
make pluto-detect-validate # Validate API versions
make pluto-detect-charts   # Scan Helm charts
```

**Decision rationale:**
- ✅ **Prevents breakage** - Catches deprecated APIs before they cause issues
- ✅ **Proactive** - Validates against target Kubernetes version (1.29)
- ✅ **Automated** - Integrated in CI/CD to prevent bad merges
- ✅ **Helm integration** - Validates Helm charts that will be deployed

---

### 6. Additional Tools - envsubst (Optional)

**What it is:**
- **envsubst** - Template substitution tool (part of `gettext`)
- Replaces environment variables in text files: `${VAR}` → actual value

**Why we considered it:**
- ✅ Useful for parameterizing test manifests
- ✅ Can be used with KUTTL or kind for E2E tests
- ✅ Helps with test environment configuration

**When to use:**
- ⬜ Parameterizing test Kubernetes manifests (if needed)
- ⬜ Configuring test environments
- ⬜ CI/CD pipelines for different environments

**Implementation:**
- ⬜ Optional - Available if needed for KUTTL or kind tests
- ⬜ Included in devbox.json
- ⬜ Makefile target: `make test-integration-envsubst` (demonstration)

**Decision rationale:**
- ✅ **Available but optional** - Included for potential future use with KUTTL
- ✅ **Devbox integration** - Available in reproducible development environment
- ✅ **Not required** - envtest doesn't need it (uses Go code directly)

---

## Complete Testing Strategy Implementation

### Testing Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱    ╲
                 ╱ E2E  ╲        ← KUTTL (Optional) ⬜
                ╱ Tests  ╲       1. Helm chart deployment
               ╱──────────╲      2. Complete workflows
              ╱            ╲     3. Upgrade testing
             ╱ Integration  ╲    ⚠️ Slower (10-30s)
            ╱   Tests        ╲   ✅ Real cluster
           ╱──────────────────╲  Run: make kuttl-test
          ╱                    ╲
         ╱   envtest ✅         ╲
        ╱────────────────────────╲
       ╱                            ╲
      ╱      Unit Tests              ╲
     ╱────────────────────────────────╲
    ╱  Ginkgo/Gomega ✅                  ╲
   ╱────────────────────────────────────╲
  ╱  Counter, Config, Parsing          ╲
 ╱────────────────────────────────────────╲
╱────────────────────────────────────────────╲
  ✅ Fast (ms)     ✅ Fast (1-3s)    ⚠️ Slower (10-30s)
  ✅ No cluster    ✅ API server     ✅ Full cluster

Static Analysis Layer (Parallel):
╱─────────────────────────────────────────────╲
╱  kubesec ✅  |  Pluto ✅                    ╲
╱  Security    |  API Version Validation      ╲
╱─────────────────────────────────────────────╲
```

### Test Coverage Matrix

| Component | Unit Tests | Integration Tests | E2E Tests | Static Analysis | Status |
|-----------|------------|-------------------|-----------|-----------------|--------|
| Counter Logic | ✅ Ginkgo | ❌ N/A | ❌ N/A | ❌ N/A | ✅ Done |
| Config Validation | ✅ Ginkgo | ❌ N/A | ❌ N/A | ❌ N/A | ✅ Done |
| Parsing Functions | ✅ Ginkgo | ❌ N/A | ❌ N/A | ❌ N/A | ✅ Done |
| Controller Logic | ✅ Ginkgo (mocks) | ✅ envtest | ⬜ KUTTL (optional) | ❌ N/A | ✅ Done |
| Informer Syncing | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ❌ N/A | ✅ Done |
| Discovery API | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ❌ N/A | ✅ Done |
| Resource Counting | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ❌ N/A | ✅ Done |
| Namespace Filtering | ❌ N/A | ✅ envtest | ⬜ KUTTL (optional) | ❌ N/A | ✅ Done |
| Helm Chart Templates | ❌ N/A | ❌ N/A | ⬜ KUTTL (optional) | ✅ kubesec + Pluto | ✅ Done |
| Security | ❌ N/A | ❌ N/A | ❌ N/A | ✅ kubesec | ✅ Done |
| API Versions | ❌ N/A | ❌ N/A | ❌ N/A | ✅ Pluto | ✅ Done |

## Decision Rationale Summary

### Why Ginkgo/Gomega for Unit Tests?

1. **BDD-style syntax** - `Describe` and `It` blocks are readable and maintainable
2. **Rich matchers** - Gomega provides extensive assertion capabilities
3. **Fast execution** - Runs in milliseconds, perfect for fast feedback
4. **Industry standard** - Widely used in Kubernetes ecosystem
5. **No dependencies** - Doesn't require cluster or external services

### Why envtest for Integration Tests?

1. **Real API server** - Essential for dynamic client, discovery, and informers
2. **Faster than kind** - 1-3 seconds vs 10-30 seconds
3. **No Docker** - Works in CI/CD without Docker dependency
4. **Perfect for controllers** - Designed specifically for controller testing
5. **Kubebuilder standard** - Recommended approach for controller-runtime projects

### Why kubesec for Security?

1. **Industry standard** - Widely used for Kubernetes security scanning
2. **Helm integration** - Scans rendered templates (actual deployed resources)
3. **Comprehensive** - Covers security contexts, RBAC, network policies
4. **Configurable** - Minimum score enforcement (default: 90)
5. **CI/CD friendly** - Easy integration in automated pipelines

### Why Pluto for API Validation?

1. **Proactive detection** - Finds deprecated APIs before cluster upgrade
2. **Target version support** - Validates against specific Kubernetes versions
3. **Helm integration** - Scans Helm charts and templates
4. **Prevents breakage** - Catches issues early before deployment
5. **CI/CD integration** - Automated validation in pipelines

### Why KUTTL is Optional?

1. **envtest covers controller logic** - Integration tests already validate controller functionality
2. **Slower execution** - 10-30 seconds vs 1-3 seconds for envtest
3. **Requires cluster** - Needs kind or existing cluster
4. **Static analysis covers Helm** - kubesec and Pluto validate Helm templates
5. **Structure created** - Ready to use when Helm chart E2E testing is needed

## Running All Tests

### Quick Commands

```bash
# Unit tests only (fast, no cluster)
make test-unit

# Integration tests (fast, API server in-process)
make test-integration

# All tests (unit + integration)
make test

# Security scan
make security-scan

# API version validation
make pluto-detect-validate

# E2E tests (optional, requires cluster)
make kuttl-install
make kuttl-test-kind  # Auto-creates kind cluster

# Full CI pipeline
make ci  # Includes: fmt, vet, lint, test, build, helm-lint, security-scan, pluto-detect-validate
```

### Test Execution Time

| Test Type | Typical Time | When to Run |
|-----------|--------------|-------------|
| Unit Tests | < 1 second | Every code change |
| Integration Tests | 1-3 seconds | Before commit |
| E2E Tests (KUTTL) | 10-30 seconds | Before release, CI/CD |
| Security Scan (kubesec) | < 5 seconds | Before commit, CI/CD |
| API Validation (Pluto) | < 5 seconds | Before commit, CI/CD |

## Implementation Files

### Unit Tests ✅
- `internal/counter/counter_test.go` - Counter logic tests
- `internal/config/config_test.go` - Config validation tests
- `main_test.go` - Main helper function tests
- `internal/controller/controller_test.go` - Controller unit tests with mocks

### Integration Tests ✅
- `internal/controller/suite_test.go` - Test suite setup (envtest)
- `internal/controller/controller_envtest_test.go` - Integration tests

### E2E Tests ⬜ (Structure Created)
- `tests/helm-deployment/` - Helm chart deployment tests
- `tests/controller-functionality/` - Controller E2E tests
- `tests/namespace-filtering/` - Namespace filtering tests
- `tests/kuttl-test.yaml` - KUTTL configuration

### Static Analysis ✅
- Makefile targets for kubesec and Pluto
- CI/CD integration in GitHub Actions
- Configuration in Makefile and `.github/workflows/ci.yml`

## Best Practices

### 1. Write Unit Tests First
✅ **Start with unit tests** - Fast, catches logic errors quickly

### 2. Add Integration Tests for API Interactions
✅ **Add integration tests** when you need real API server behavior

### 3. Use Static Analysis for Helm Charts
✅ **Always run** kubesec and Pluto on Helm templates before deployment

### 4. E2E Tests for Complete Workflows (Optional)
✅ **Add E2E tests** when you need Helm chart or upgrade testing

### 5. Run Everything in CI/CD
✅ **Automate all tests** - Unit, integration, security, API validation in CI/CD

## Next Steps

### ✅ Completed
1. ✅ Unit tests with Ginkgo/Gomega
2. ✅ Integration tests with envtest
3. ✅ Security scanning with kubesec
4. ✅ API version validation with Pluto
5. ✅ KUTTL test structure created

### ⬜ Optional
1. ⬜ Implement KUTTL tests when Helm chart E2E testing is needed
2. ⬜ Add KUTTL to CI/CD pipeline if E2E testing becomes required

## References

- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - Quick decision guide
- [envtest Explained](./ENVTEST_EXPLAINED.md) - Detailed envtest guide
- [KUTTL Usage](./KUTTL_USAGE.md) - KUTTL E2E testing guide
- [KUTTL FAQ](./KUTTL_FAQ.md) - KUTTL frequently asked questions
- [Kubebuilder Testing Guide](https://book.kubebuilder.io/cronjob-tutorial/writing-tests) - Official envtest documentation
