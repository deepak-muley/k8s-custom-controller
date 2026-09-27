# Understanding envtest for Controller Testing

## What is envtest?

**envtest** is a testing framework from `sigs.k8s.io/controller-runtime/pkg/envtest` that provides a **real Kubernetes API server** for testing controllers without requiring a full Kubernetes cluster.

Reference: [Kubebuilder Testing Guide](https://book.kubebuilder.io/cronjob-tutorial/writing-tests)

## How envtest Works

```
┌─────────────────────────────────────────────────────────┐
│                    Your Test Code                       │
│  • Ginkgo/Gomega test suite                             │
│  • Creates resources using dynamic client                │
│  • Starts controller                                     │
│  • Validates behavior                                    │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   envtest.Environment                   │
│  • Starts etcd (database)                                │
│  • Starts Kubernetes API server                          │
│  • Provides REST config                                  │
│  • Handles cleanup                                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Real Kubernetes API Server                  │
│  • Full API server functionality                         │
│  • Discovery API                                         │
│  • Resource CRUD operations                              │
│  • Watch/Informer support                                │
│  • etcd backend                                          │
└─────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Test Suite Setup (`suite_test.go`)

```go
var (
    ctx       context.Context
    cancel    context.CancelFunc
    testEnv   *envtest.Environment  // Manages API server lifecycle
    cfg       *rest.Config           // REST config to connect to test API server
    k8sClient client.Client          // Client for creating test resources
)

func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
    // Start test environment (API server + etcd)
    testEnv = &envtest.Environment{
        CRDDirectoryPaths:     []string{...},  // Optional: if you have CRDs
        ErrorIfCRDPathMissing: false,
    }
    
    cfg, err := testEnv.Start()  // Starts API server and returns REST config
    Expect(err).NotTo(HaveOccurred())
    
    // Create clients for testing
    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
    // Cleanup: Stop API server and etcd
    err := testEnv.Stop()
    Expect(err).NotTo(HaveOccurred())
})
```

### 2. Integration Tests (`controller_envtest_test.go`)

```go
var _ = Describe("GVKController with envtest", func() {
    It("should discover resources correctly", func() {
        // Create controller with REST config from envtest
        controllerCfg := &config.Config{...}
        controller, err := NewGVKController(controllerCfg, cfg, scheme)
        // cfg is the REST config from envtest - points to test API server
        
        // Test discovery API - this uses REAL API server
        resources, err := controller.discoveryClient.ServerResourcesForGroupVersion("apps/v1")
        // Real discovery! Returns actual resource definitions
        
        // Test informer setup - this uses REAL API server
        controller.Start(ctx)
        // Informers will sync with real API server cache
    })
    
    It("should count resources when they are created", func() {
        // Create a deployment using dynamic client
        deployment := &unstructured.Unstructured{...}
        controller.dynamicClient.Resource(gvr).Create(ctx, deployment, ...)
        // This creates resource in REAL API server
        
        // Wait for controller informer to process the event
        Eventually(func() int {
            return controller.counter.GetCount(deploymentGVK)
        }).Should(Equal(1))
        // Real watch event → Informer → Event handler → Counter increment!
    })
})
```

## Why envtest Works Perfectly for Your Controller

### Your Controller Needs:

1. **Discovery API** ✅
   ```go
   discoveryClient.ServerResourcesForGroupVersion("apps/v1")
   ```
   - Fake clients: ❌ Can't properly fake discovery
   - envtest: ✅ Real API server with full discovery support

2. **Dynamic Client** ✅
   ```go
   dynamicClient.Resource(gvr).Create(...)
   ```
   - Fake clients: ❌ Limited dynamic client support
   - envtest: ✅ Full dynamic client support

3. **Informers with Watch** ✅
   ```go
   informerFactory.ForResource(gvr).Informer()
   ```
   - Fake clients: ❌ Informers don't work properly
   - envtest: ✅ Real watch streams and cache syncing

4. **Resource Events (Add/Update/Delete)** ✅
   ```go
   informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
       AddFunc: func(obj interface{}) { counter.Increment(...) }
   })
   ```
   - Fake clients: ❌ Events don't flow correctly
   - envtest: ✅ Real events from API server

## What envtest Provides vs What You Don't Need

| Component | Provided by envtest? | Needed for Your Controller? |
|-----------|---------------------|----------------------------|
| API Server | ✅ Yes | ✅ Yes (Essential!) |
| etcd | ✅ Yes | ✅ Yes (Storage) |
| Discovery API | ✅ Yes | ✅ Yes (Resource discovery) |
| Watch API | ✅ Yes | ✅ Yes (Informers) |
| Resource CRUD | ✅ Yes | ✅ Yes (Create/Delete tests) |
| kubelet | ❌ No | ❌ No (Not needed) |
| Pod Scheduling | ❌ No | ❌ No (Not needed) |
| Node Components | ❌ No | ❌ No (Not needed) |

**Result:** envtest provides exactly what you need, nothing more, nothing less!

## Comparison: envtest vs Other Approaches

### envtest vs Fake Clients

| Aspect | Fake Clients | envtest |
|--------|-------------|---------|
| **Discovery API** | ❌ Hard to fake | ✅ Real API server |
| **Dynamic Clients** | ❌ Limited support | ✅ Full support |
| **Informers** | ❌ Don't work | ✅ Real informers |
| **Cache Syncing** | ❌ Can't test | ✅ Real syncing |
| **Speed** | ⚡ Very fast | ✅ Fast (~1-3s) |
| **Realism** | ❌ Simulated | ✅ Real API |

**Verdict:** envtest is better for your controller because it provides real API server functionality.

### envtest vs kind/testcontainers

| Aspect | kind/testcontainers | envtest |
|--------|---------------------|---------|
| **Startup Time** | ⚠️ 10-30 seconds | ✅ 1-3 seconds |
| **Docker Required** | ❌ Yes | ✅ No |
| **Resource Usage** | ⚠️ High | ✅ Low |
| **API Server** | ✅ Yes | ✅ Yes |
| **kubelet/Pods** | ✅ Yes | ❌ No (but not needed!) |
| **CI/CD Setup** | ⚠️ Complex | ✅ Simple |

**Verdict:** envtest is better for your controller because you don't need kubelet/pod scheduling, just API server functionality.

## How to Use envtest in Your Tests

### 1. Install Dependencies

```bash
go get sigs.k8s.io/controller-runtime/pkg/envtest
go get sigs.k8s.io/controller-runtime/pkg/client
```

### 2. Set Up Test Suite

See `internal/controller/suite_test.go` for complete setup.

### 3. Write Tests

```go
var _ = Describe("MyController", func() {
    It("should work with real API server", func() {
        // Create controller using cfg (REST config from envtest)
        controller, err := NewGVKController(config, cfg, scheme)
        
        // Start controller - it connects to envtest API server
        controller.Start(ctx)
        
        // Create resources - they go to real API server
        controller.dynamicClient.Resource(gvr).Create(...)
        
        // Controller informer receives real watch events!
        Eventually(func() int {
            return controller.counter.GetCount(gvk)
        }).Should(Equal(1))
    })
})
```

### 4. Run Tests

```bash
# Run all tests (includes envtest integration tests)
make test

# Or with ginkgo
ginkgo -v ./internal/controller/

# Or with go test
go test -v ./internal/controller/...
```

## Understanding envtest Configuration

```go
testEnv = &envtest.Environment{
    // Optional: Path to CRD manifests if you have custom resources
    CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
    
    // Don't fail if CRD path doesn't exist (if you don't have CRDs)
    ErrorIfCRDPathMissing: false,
    
    // Optional: Specify Kubernetes version
    // BinaryAssetsDirectory: "/path/to/kubernetes/binaries",
    
    // Optional: Configure API server flags
    // AttachControlPlaneOutput: true,  // Show API server logs
}
```

## Benefits for Your Project

1. **Fast Feedback** - Tests run in 1-3 seconds vs 10-30 seconds with kind
2. **No Docker Required** - Works everywhere, including CI/CD
3. **Real API Semantics** - Tests actual Kubernetes API behavior
4. **Perfect for Controllers** - Designed specifically for controller testing
5. **Industry Standard** - Used by Kubebuilder and controller-runtime projects
6. **Easy CI/CD** - Works in GitHub Actions without additional setup

## Common Patterns

### Pattern 1: Testing Discovery

```go
It("should discover resources", func() {
    resources, err := discoveryClient.ServerResourcesForGroupVersion("apps/v1")
    // Real discovery API call!
    Expect(err).NotTo(HaveOccurred())
    Expect(resources.APIResources).To(ContainElement(...))
})
```

### Pattern 2: Testing Resource Counting

```go
It("should count created resources", func() {
    // Create resource in real API server
    deployment := &unstructured.Unstructured{...}
    controller.dynamicClient.Resource(gvr).Create(ctx, deployment, ...)
    
    // Wait for informer event handler
    Eventually(func() int {
        return controller.counter.GetCount(gvk)
    }, timeout, interval).Should(Equal(1))
})
```

### Pattern 3: Testing Namespace Filtering

```go
It("should filter by namespace", func() {
    // Create controller watching specific namespace
    cfg := &config.Config{Namespace: "test-ns", ...}
    controller, _ := NewGVKController(cfg, restConfig, scheme)
    
    // Create resources in different namespaces
    controller.dynamicClient.Resource(gvr).Namespace("test-ns").Create(...)
    controller.dynamicClient.Resource(gvr).Namespace("other-ns").Create(...)
    
    // Only watched namespace should be counted
    Eventually(func() int {
        return controller.counter.GetCount(gvk)
    }).Should(Equal(1))  // Only one from test-ns
})
```

## Summary

**envtest is the perfect choice for your controller because:**
- ✅ Provides real Kubernetes API server (not mocked)
- ✅ Supports dynamic clients and discovery API
- ✅ Informers work correctly with real watch events
- ✅ Fast (1-3 seconds startup vs 10-30 for kind)
- ✅ No Docker required
- ✅ Industry standard (Kubebuilder pattern)
- ✅ Perfect for testing controllers

**Your controller needs exactly what envtest provides - nothing more, nothing less!**
