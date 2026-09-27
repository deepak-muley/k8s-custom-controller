# E2E Testing Approaches Comparison

## Overview

This document compares different approaches to End-to-End (E2E) testing for Kubernetes controllers, specifically comparing:
- **KUTTL** (YAML-based declarative testing)
- **Golang E2E with kind** (Code-based testing)
- **Cluster API E2E Framework** (Framework-based testing)

Reference: [Cluster API E2E Testing Guide](https://cluster-api.sigs.k8s.io/developer/core/e2e)

## Quick Comparison Table

| Aspect | KUTTL | Golang E2E with kind | Cluster API E2E Framework |
|--------|-------|---------------------|---------------------------|
| **Language** | YAML + Bash scripts | Go | Go (framework-based) |
| **Approach** | Declarative | Imperative | Framework-driven |
| **Test Format** | YAML manifests | Go code | Go code with framework |
| **Setup Complexity** | Low | Medium | High |
| **Learning Curve** | Low (YAML) | Medium (Go) | High (Framework) |
| **Flexibility** | Medium | High | Very High |
| **Maintainability** | High (simple YAML) | Medium | High (reusable patterns) |
| **Cluster Management** | Auto (kind) or existing | Manual (kind) | Framework-managed |
| **Helm Chart Testing** | ✅ Excellent | ✅ Good | ⚠️ Complex |
| **Portability** | High | Medium | Very High |
| **CI/CD Integration** | ✅ Easy | ✅ Easy | ✅ Excellent |
| **Best For** | Helm charts, simple E2E | Custom controllers, full control | Large projects, reusable patterns |

## Detailed Comparison

### 1. KUTTL (YAML-based Declarative Testing)

**What it is:**
- Declarative testing framework using YAML manifests
- Tests Kubernetes resources by creating them and asserting state
- Uses scripts for setup (e.g., Helm chart installation)
- Reference: [KUTTL GitHub](https://github.com/kudobuilder/kuttl)

**Advantages:**
- ✅ **Simple syntax** - Just YAML manifests, easy to understand
- ✅ **Excellent for Helm charts** - Can install and test Helm charts easily
- ✅ **Declarative** - Tests describe desired state, not how to achieve it
- ✅ **No Go knowledge required** - Works with YAML and bash scripts
- ✅ **Auto cluster management** - Can create/destroy kind clusters automatically
- ✅ **Fast feedback** - Easy to see what's being tested by reading YAML
- ✅ **Good for CI/CD** - Simple to integrate in pipelines

**Disadvantages:**
- ⚠️ **Limited programmability** - Hard to do complex logic in YAML
- ⚠️ **Script dependencies** - Requires bash scripts for Helm installation
- ⚠️ **Less flexible** - Difficult to handle complex test scenarios
- ⚠️ **State verification** - Harder to verify internal controller state

**When to use:**
- ✅ Testing Helm chart deployments
- ✅ Simple E2E workflows
- ✅ Teams with less Go expertise
- ✅ When you need quick, readable tests

**Example Structure:**
```yaml
# tests/helm-deployment/00-install-helm.sh
#!/bin/bash
helm install test-controller ./helm/k8s-custom-controller \
  --namespace kuttl-test-helm \
  --set controller.namespace=test

# tests/helm-deployment/01-assert-resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1
```

**Implementation in this project:**
- ✅ Test structure created in `tests/helm-deployment/`
- ✅ Script for Helm chart installation: `tests/helm-deployment/00-install-helm.sh`
- ✅ Assertion files for resource verification
- ✅ Makefile targets: `make kuttl-test`, `make kuttl-test-kind`

---

### 2. Golang E2E with kind (Code-based Testing)

**What it is:**
- Write E2E tests in Go using Ginkgo/Gomega
- Manually manage kind cluster lifecycle
- Use `controller-runtime` client and kubectl programmatically
- Similar to what we use for integration tests, but with full cluster

**Advantages:**
- ✅ **Full Go power** - Can use all Go features and libraries
- ✅ **Complete control** - Full control over test execution
- ✅ **Easy debugging** - Can set breakpoints, use debugger
- ✅ **Reusable code** - Can create helper functions and utilities
- ✅ **Type safety** - Compile-time checks for Kubernetes resources
- ✅ **Integration with existing tests** - Uses same testing framework (Ginkgo)

**Disadvantages:**
- ⚠️ **More code** - Requires writing more Go code
- ⚠️ **Manual cluster management** - Need to create/destroy kind clusters manually
- ⚠️ **Helm integration** - Need to exec `helm` command or use Helm Go SDK
- ⚠️ **More complex setup** - Requires understanding kind, Go, and Kubernetes clients

**When to use:**
- ✅ When you need complex test logic
- ✅ When you want full control over test execution
- ✅ When you have Go expertise
- ✅ When you need to test programmatic scenarios

**Example Structure:**
```go
// e2e/helm_chart_test.go
var _ = Describe("Helm Chart E2E", func() {
    var kindCluster *kind.Cluster
    
    BeforeEach(func() {
        // Create kind cluster
        kindCluster = kind.NewCluster("e2e-test")
        Expect(kindCluster.Create()).To(Succeed())
        
        // Build and load Docker image
        Expect(buildDockerImage()).To(Succeed())
        Expect(kindCluster.LoadImage("k8s-custom-controller:test")).To(Succeed())
        
        // Install Helm chart
        Expect(installHelmChart(kindCluster)).To(Succeed())
    })
    
    AfterEach(func() {
        kindCluster.Delete()
    })
    
    It("should deploy Helm chart correctly", func() {
        // Verify deployment exists
        deployment := &appsv1.Deployment{}
        Expect(k8sClient.Get(ctx, client.ObjectKey{
            Name: "k8s-custom-controller",
            Namespace: "kuttl-test-helm",
        }, deployment)).To(Succeed())
        
        Expect(deployment.Status.ReadyReplicas).To(Equal(int32(1)))
    })
})
```

**Implementation in this project:**
- ⬜ Not implemented (KUTTL preferred for Helm chart testing)
- ✅ Structure available if needed
- ✅ Can be added to `e2e/` directory

---

### 3. Cluster API E2E Framework

**What it is:**
- Sophisticated E2E testing framework from Cluster API project
- Provides helper methods for common E2E tasks
- Designed for testing Kubernetes controllers and operators
- Reference: [Cluster API E2E Guide](https://cluster-api.sigs.k8s.io/developer/core/e2e)

**Advantages:**
- ✅ **Mature framework** - Battle-tested in Cluster API project
- ✅ **Rich helpers** - Many helper methods for common tasks
- ✅ **Portable tests** - Tests can run with different infrastructure providers
- ✅ **Best practices** - Follows Kubernetes E2E testing best practices
- ✅ **Cluster lifecycle** - Handles cluster creation/management automatically
- ✅ **Resource cleanup** - Automatic cleanup of test resources
- ✅ **Log collection** - Automatic log collection from controllers
- ✅ **Configurable** - E2E config files for different environments

**Disadvantages:**
- ❌ **High complexity** - Steeper learning curve
- ❌ **Overkill for simple projects** - Designed for large, complex projects
- ❌ **Framework dependency** - Depends on Cluster API test framework
- ❌ **Less flexibility** - Framework patterns may not fit all use cases
- ❌ **More setup** - Requires E2E config files, framework initialization

**When to use:**
- ✅ Large, complex Kubernetes projects
- ✅ Projects with multiple infrastructure providers
- ✅ When you need portable tests across providers
- ✅ When you want battle-tested patterns from Cluster API
- ❌ **Not recommended** for simple custom controllers

**Example Structure (Based on Cluster API Patterns):**
```go
// Based on Cluster API E2E patterns
// Reference: https://cluster-api.sigs.k8s.io/developer/core/e2e

var (
    ctx        = context.TODO()
    configPath = "/path/to/e2e-config.yaml"
    e2eConfig  *clusterctl.E2EConfig
    cluster    *framework.ClusterProxy
)

var _ = Describe("Controller E2E", Ordered, func() {
    BeforeAll(func() {
        // 1. Load E2E config file
        // The E2E config defines:
        // - Providers to install (Cluster API core, infrastructure providers)
        // - Variables for clusterctl (Kubernetes version, CNI, etc.)
        // - Intervals for wait operations (timeouts)
        // - Images to load into kind cluster
        e2eConfig = clusterctl.LoadE2EConfig(ctx, configPath)
        
        // 2. Create management cluster (kind or existing)
        // Cluster API framework manages cluster lifecycle
        cluster = framework.CreateManagementCluster(ctx, framework.CreateManagementClusterInput{
            Config: e2eConfig,
        })
        
        // 3. Install providers in management cluster
        // Runs "clusterctl init" with providers from config
        // Waits for controllers to be running
        // Creates log watchers for all providers
        framework.InitManagementClusterAndWatchControllerLogs(ctx, framework.InitManagementClusterAndWatchControllerLogsInput{
            ClusterProxy: cluster,
            Config: e2eConfig,
        })
    })
    
    It("should deploy and test controller", func() {
        // 1. Create test namespace
        // Framework provides helper for namespace creation
        namespace := framework.CreateNamespaceAndWatchEvents(ctx, framework.CreateNamespaceAndWatchEventsInput{
            ClusterProxy: cluster,
            Name: "test-namespace",
        })
        
        // 2. Create resources using cluster templates
        // Uses clusterctl to generate cluster manifests from templates
        // Templates are provided via E2E config file
        // This makes tests portable across infrastructure providers
        framework.ApplyClusterTemplateAndWait(ctx, framework.ApplyClusterTemplateAndWaitInput{
            ConfigProxy: cluster,
            Config: e2eConfig,
            ClusterName: "test-cluster",
            Namespace: namespace,
        })
        
        // 3. Wait for infrastructure to be provisioned
        // Framework provides helpers for waiting on Cluster API resources
        framework.WaitForClusterToProvision(ctx, framework.WaitForClusterToProvisionInput{
            ClusterProxy: cluster,
            Cluster: cluster,
        })
        
        // 4. Verify resources
        // Framework provides Get and Wait methods for resources
        framework.WaitForKubeadmControlPlaneMachinesToExist(ctx, framework.WaitForKubeadmControlPlaneMachinesToExistInput{
            ClusterProxy: cluster,
            Cluster: cluster,
        })
    })
    
    AfterAll(func() {
        // Framework automatically collects:
        // - All logs from Cluster API controllers
        // - All Cluster API/Kubernetes objects
        // - Cleans up infrastructure resources
        framework.DumpClusterResourcesAndCleanup(ctx, framework.DumpClusterResourcesAndCleanupInput{
            ClusterProxy: cluster,
            Config: e2eConfig,
        })
    })
})
```

**E2E Config File Example (Cluster API Pattern):**
```yaml
# e2e-config.yaml
# Reference: https://cluster-api.sigs.k8s.io/developer/core/e2e
# This config defines the management cluster setup

# 1. Define providers to install in management cluster
providers:
- name: "cluster-api"
  type: "CoreProvider"
  versions:
  - name: "v1.6.0"
    value: "file:///path/to/cluster-api-manifests"  # Built from sources or remote
  # Can define multiple versions for upgrade testing
  # Additional files can be added (e.g., cluster-templates.yaml)

- name: "aws"  # Infrastructure provider
  type: "InfrastructureProvider"
  versions:
  - name: "v2.3.0"
    value: "file:///path/to/capa-manifests"
  additionalFiles:
  - file: "cluster-templates.yaml"  # Cluster templates for different scenarios

# 2. Define variables for clusterctl operations
variables:
  KUBERNETES_VERSION: "v1.29.0"  # Used in clusterctl generate cluster
  CNI: "calico"
  AWS_REGION: "us-west-2"
  AWS_SSH_KEY_NAME: "test-key"
  # These variables are used when running clusterctl init/generate

# 3. Define intervals for wait operations
intervals:
  default/wait-cluster: "10m"
  default/wait-deployment: "5m"
  default/wait-control-plane: "10m"
  # Tests use these intervals via framework.GetIntervals()

# 4. Define images to load into kind cluster (if using kind)
images:
- name: "k8s-custom-controller"
  loadBehavior: "mustLoad"  # or "tryLoad" or "kindLoad"
  # Framework will load these images into kind before tests
```

**Implementation in this project:**
- ❌ Not implemented (overkill for this project)
- ⚠️ Would require significant framework setup
- ❌ Not recommended for simple custom controller

---

## Detailed Feature Comparison

### Helm Chart Testing

| Feature | KUTTL | Golang E2E | Cluster API Framework |
|---------|-------|------------|----------------------|
| **Helm Install** | ✅ Script-based (easy) | ⚠️ Exec or SDK (medium) | ⚠️ Manual (complex) |
| **Helm Values** | ✅ Easy (script args) | ⚠️ Need to construct | ⚠️ Manual |
| **Upgrade Testing** | ✅ Easy (script) | ⚠️ Manual | ⚠️ Manual |
| **Template Validation** | ❌ Not direct | ✅ Can render templates | ✅ Can render templates |
| **Multi-version Testing** | ✅ Easy | ⚠️ Manual | ✅ Supported |

**Winner: KUTTL** - Best for Helm chart testing with simple scripts

### Cluster Management

| Feature | KUTTL | Golang E2E | Cluster API Framework |
|---------|-------|------------|----------------------|
| **Auto Create** | ✅ Built-in (startKind) | ⚠️ Manual | ✅ Framework-managed |
| **Auto Destroy** | ✅ Built-in | ⚠️ Manual (AfterSuite) | ✅ Built-in |
| **Image Loading** | ⚠️ Manual | ✅ Manual (kind API) | ✅ Framework helpers |
| **Multi-cluster** | ❌ Single cluster | ⚠️ Manual | ✅ Supported |

**Winner: Cluster API Framework** - Best cluster management, but KUTTL is simpler for single cluster

### Test Maintainability

| Feature | KUTTL | Golang E2E | Cluster API Framework |
|---------|-------|------------|----------------------|
| **Readability** | ✅ Very high (YAML) | ⚠️ Medium (Go code) | ⚠️ Medium (Go + framework) |
| **Simplicity** | ✅ Very simple | ⚠️ More complex | ❌ Most complex |
| **Debugging** | ⚠️ Limited | ✅ Full Go debugging | ✅ Full Go debugging |
| **Reusability** | ⚠️ Limited | ✅ High (functions) | ✅ Very high (patterns) |

**Winner: KUTTL** - Simplest and most readable, but Golang E2E offers more flexibility

### CI/CD Integration

| Feature | KUTTL | Golang E2E | Cluster API Framework |
|---------|-------|------------|----------------------|
| **Setup Time** | ✅ Fast (< 1 min) | ⚠️ Medium (2-5 min) | ❌ Slow (5-10 min) |
| **Dependencies** | ✅ Minimal | ⚠️ Go + kind | ❌ Framework + dependencies |
| **Parallel Execution** | ✅ Supported | ✅ Supported | ✅ Supported |
| **Artifact Collection** | ⚠️ Limited | ✅ Full control | ✅ Comprehensive |

**Winner: KUTTL** - Fastest setup, easiest CI/CD integration

---

## Real-World Examples

### Example 1: Testing Helm Chart Installation

#### KUTTL Approach ✅ Recommended

```bash
# tests/helm-deployment/00-install-helm.sh
helm install test-controller ./helm/k8s-custom-controller \
  --namespace kuttl-test-helm \
  --set controller.namespace=test \
  --set controller.gvks="apps/v1/Deployment"

# tests/helm-deployment/01-assert-resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
  namespace: kuttl-test-helm
status:
  readyReplicas: 1
```

**Pros:** Simple, readable, easy to maintain  
**Cons:** Limited to YAML assertions

#### Golang E2E Approach

```go
var _ = Describe("Helm Chart E2E", func() {
    It("should install Helm chart", func() {
        // Install chart
        cmd := exec.Command("helm", "install", "test-controller", 
            "./helm/k8s-custom-controller",
            "--namespace", "kuttl-test-helm",
            "--set", "controller.namespace=test")
        Expect(cmd.Run()).To(Succeed())
        
        // Wait for deployment
        Eventually(func() error {
            deployment := &appsv1.Deployment{}
            return k8sClient.Get(ctx, client.ObjectKey{
                Name: "k8s-custom-controller",
                Namespace: "kuttl-test-helm",
            }, deployment)
        }).Should(Succeed())
        
        // Assert state
        deployment := &appsv1.Deployment{}
        k8sClient.Get(ctx, client.ObjectKey{
            Name: "k8s-custom-controller",
            Namespace: "kuttl-test-helm",
        }, deployment)
        Expect(deployment.Status.ReadyReplicas).To(Equal(int32(1)))
    })
})
```

**Pros:** Full Go control, type-safe  
**Cons:** More code, manual Helm exec

#### Cluster API Framework Approach

```go
var _ = Describe("Helm Chart E2E", func() {
    It("should install Helm chart", func() {
        // Framework manages cluster, but Helm install is manual
        // Need to use framework helpers for resource creation
        // More complex setup required
    })
})
```

**Pros:** Reusable patterns, comprehensive  
**Cons:** Overkill, complex setup

**Winner: KUTTL** - Simplest and most appropriate for Helm chart testing

---

### Example 2: Testing Controller Functionality

#### KUTTL Approach

```yaml
# Create test resources
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-controller
# ... deployment spec

# Verify controller counted (requires checking logs/metrics)
# Harder to verify internal state
```

**Pros:** Simple resource creation  
**Cons:** Hard to verify controller internal state

#### Golang E2E Approach ✅ Recommended for Complex Logic

```go
It("should count resources correctly", func() {
    // Create test deployment
    deployment := &appsv1.Deployment{...}
    Expect(k8sClient.Create(ctx, deployment)).To(Succeed())
    
    // Wait for controller to process
    Eventually(func() error {
        // Query controller state via API/metrics/logs
        return verifyControllerCount(deployment, 1)
    }).Should(Succeed())
    
    // Delete deployment
    Expect(k8sClient.Delete(ctx, deployment)).To(Succeed())
    
    // Verify count decreased
    Eventually(func() error {
        return verifyControllerCount(deployment, 0)
    }).Should(Succeed())
})
```

**Pros:** Full control, can verify internal state  
**Cons:** More code required

**Winner: Golang E2E** - Better for complex controller logic testing

---

## Recommendation for This Project

### ✅ Use KUTTL for:

1. **Helm Chart E2E Testing** ⭐ **Primary Use Case**
   - Install Helm chart via script (`00-install-helm.sh`)
   - Assert resources exist and are correct
   - Test upgrade scenarios
   - Simple, readable, maintainable

2. **Simple E2E Workflows**
   - Deploy → Verify → Cleanup
   - Resource creation and validation
   - Quick feedback in CI/CD

### ✅ Use Golang E2E for:

1. **Complex Controller Logic Testing**
   - Testing internal controller state
   - Complex scenarios requiring Go logic
   - Programmatic resource manipulation
   - Integration with existing Ginkgo tests

2. **When You Need Full Control**
   - Custom test scenarios
   - Complex debugging requirements
   - Type-safe resource manipulation

### ❌ Don't Use Cluster API Framework for:

1. **This Project** - Overkill for a simple custom controller
2. **Simple Use Cases** - Too much complexity for basic E2E testing
3. **Helm Chart Testing** - Framework doesn't have built-in Helm support

---

## Implementation Status

### ✅ Implemented (This Project)

1. **KUTTL** - Helm chart E2E testing
   - ✅ Test structure created
   - ✅ Helm installation script (`tests/helm-deployment/00-install-helm.sh`)
   - ✅ Resource assertion files
   - ✅ Makefile targets (`make kuttl-test-kind`)

2. **Integration Tests (envtest)** - Controller logic testing
   - ✅ Full implementation with envtest
   - ✅ Tests discovery, informers, counting
   - ✅ Fast feedback (1-3 seconds)

### ⬜ Optional (Can Be Added)

1. **Golang E2E with kind** - Complex scenarios
   - ⬜ Can be added to `e2e/` directory if needed
   - ⬜ Would complement KUTTL for complex scenarios

2. **Cluster API Framework** - Not recommended
   - ❌ Overkill for this project
   - ❌ Too much complexity for simple controller

---

## Comparison Summary

| Criteria | Winner | Why |
|----------|--------|-----|
| **Helm Chart Testing** | KUTTL | Simplest, script-based installation |
| **Simple E2E Tests** | KUTTL | YAML-based, easy to read and maintain |
| **Complex Logic Testing** | Golang E2E | Full Go power, type safety |
| **Large Projects** | Cluster API Framework | Mature patterns, portability |
| **CI/CD Integration** | KUTTL | Fastest setup, minimal dependencies |
| **Learning Curve** | KUTTL | YAML is easier than Go/frameworks |
| **Maintainability** | KUTTL | Simple YAML is easier to maintain |

---

## Key Differences Summary

### Helm Chart Testing Approach

| Approach | How It Works | Pros | Cons | Best For |
|----------|--------------|------|------|----------|
| **KUTTL** | Script installs chart via `helm install`, then asserts resources | ✅ Simple, readable YAML | ⚠️ Requires script | Helm chart E2E testing |
| **Golang E2E** | Code execs `helm install` or uses Helm Go SDK | ✅ Full control, type-safe | ⚠️ More code | Complex scenarios |
| **Cluster API** | Manual Helm install or framework helpers | ✅ Reusable patterns | ❌ Complex setup | Large projects |

### Our Implementation: KUTTL with Helm Chart Install

**✅ What We Use:**
```bash
# tests/helm-deployment/00-install-helm.sh
# Full Helm chart installation via script

helm upgrade --install k8s-custom-controller ./helm/k8s-custom-controller \
  --namespace kuttl-test-helm \
  --set image.repository=k8s-custom-controller \
  --set image.tag=test \
  --wait \
  --timeout 5m

# Then verify all resources in 01-assert-resources.yaml
```

**Key Points:**
- ✅ **Full Helm Install** - Uses `helm install` (not individual YAMLs)
- ✅ **All Resources** - Installs Deployment, SA, RBAC, ConfigMap, etc. via Helm
- ✅ **Helm Values** - Tests with different `--set` flags
- ✅ **Helm Lifecycle** - Respects Helm hooks, dependencies, and lifecycle
- ✅ **Real Deployment** - Tests actual Helm deployment workflow

**Difference from Individual YAMLs:**
- ❌ **NOT** installing individual YAML files
- ❌ **NOT** testing rendered templates only
- ✅ **FULL** Helm chart installation with all templates, values, and hooks

## Cluster API E2E Framework Details

### Framework Features (from Cluster API Guide)

**Reference:** [Cluster API E2E Testing Guide](https://cluster-api.sigs.k8s.io/developer/core/e2e)

1. **E2E Config File**
   - Defines providers, versions, variables, intervals
   - Makes tests portable across infrastructure providers
   - Allows different configurations for different environments

2. **Management Cluster Setup**
   - `InitManagementClusterAndWatchControllerLogs()` - Installs providers
   - Handles cluster lifecycle automatically
   - Creates log watchers for all providers

3. **Resource Creation**
   - `ApplyClusterTemplateAndWait()` - Creates resources from templates
   - Uses cluster-templates.yaml files for portability
   - Framework methods for waiting on resources

4. **Cleanup and Log Collection**
   - `DumpClusterResourcesAndCleanup()` - Automatic cleanup
   - Collects all logs from controllers
   - Dumps all Cluster API/Kubernetes objects
   - Cleans up infrastructure resources

5. **Portable Tests**
   - Tests can run with different infrastructure providers
   - Use E2E config to swap providers
   - Tests adapt to different cluster-templates.yaml files

### Why Cluster API Framework is Overkill for This Project

1. **Designed for Cluster API** - Framework is specific to Cluster API use cases
2. **Provider Management** - We don't have infrastructure providers to manage
3. **Complexity** - Too much setup for a simple custom controller
4. **Helm Support** - Framework doesn't have built-in Helm chart testing
5. **Simple Controller** - Our controller doesn't need Cluster API patterns

**Recommendation:** Use **KUTTL** for Helm chart testing, **envtest** for controller logic. Cluster API framework is not needed.

## Helm Chart Installation Example: Full Install vs Individual YAMLs

### ✅ Full Helm Chart Install (What We Use)

**This project uses KUTTL with a script-based approach for full Helm chart installation:**

```bash
# tests/helm-deployment/00-install-helm.sh
# This script performs a FULL Helm chart installation
# It installs ALL resources via "helm install" (not individual YAMLs)

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
- ✅ Helm hooks (pre-install, post-install, etc.)
- ✅ Helm dependencies (if configured)

**Key Benefits:**
1. **Real Helm Deployment** - Tests actual Helm installation workflow
2. **All Templates** - Installs all resources defined in Helm templates
3. **Values Handling** - Tests with different values.yaml configurations
4. **Helm Lifecycle** - Respects Helm hooks and dependencies
5. **Upgrade Testing** - Can test Helm upgrades/downgrades easily

### ❌ Individual YAMLs Approach (What We DON'T Use)

**We do NOT install individual YAML files:**

```yaml
# ❌ NOT doing this:
# Installing individual YAML files manually

apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-custom-controller
# ... manually applying each resource
```

**Why we don't use individual YAMLs:**
- ❌ Doesn't test real Helm deployment workflow
- ❌ Doesn't test Helm templates and values
- ❌ Doesn't test Helm hooks and dependencies
- ❌ Doesn't test Helm lifecycle
- ❌ Harder to test upgrades/downgrades

### Comparison: Full Helm Install vs Individual YAMLs

| Aspect | Full Helm Install (✅ What We Use) | Individual YAMLs (❌ What We Don't Use) |
|--------|-------------------------------------|------------------------------------------|
| **Approach** | `helm install` command | Manual `kubectl apply` of individual files |
| **Helm Templates** | ✅ All templates applied | ❌ Need to render manually |
| **Helm Values** | ✅ `--set` flags work | ❌ Need to substitute manually |
| **Helm Hooks** | ✅ Executed | ❌ Not executed |
| **Helm Dependencies** | ✅ Resolved | ❌ Not resolved |
| **Helm Lifecycle** | ✅ Full lifecycle | ❌ No lifecycle |
| **Upgrade Testing** | ✅ Easy (`helm upgrade`) | ❌ Manual |
| **Real-World Testing** | ✅ Tests actual deployment | ❌ Tests rendered templates only |

### Implementation in This Project

**✅ What We Have:**
- ✅ `tests/helm-deployment/00-install-helm.sh` - Script that installs full Helm chart
- ✅ `tests/helm-deployment/01-assert-resources.yaml` - Asserts all resources exist
- ✅ `tests/helm-deployment/steps/` - Additional verification steps
- ✅ Makefile targets: `make kuttl-test-helm`, `make kuttl-test-kind`

**How It Works:**
1. **KUTTL executes script** - `00-install-helm.sh` runs automatically
2. **Script installs Helm chart** - `helm install` command installs entire chart
3. **All resources created** - Deployment, SA, RBAC, ConfigMap, etc. via Helm
4. **KUTTL verifies resources** - `01-assert-resources.yaml` waits for resources
5. **Additional verification** - Steps verify controller functionality

**Run Command:**
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

## References

- [KUTTL GitHub](https://github.com/kudobuilder/kuttl) - YAML-based E2E testing
- [Cluster API E2E Guide](https://cluster-api.sigs.k8s.io/developer/core/e2e) - Framework-based E2E testing (reference)
- [Kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [Ginkgo](https://onsi.github.io/ginkgo/) - BDD testing framework for Go
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - Quick decision guide
- [Testing Strategy](./TESTING_STRATEGY.md) - Complete testing strategy
- [KUTTL Usage](./KUTTL_USAGE.md) - Detailed KUTTL usage guide
