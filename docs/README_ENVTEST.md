# envtest Integration Tests - Quick Reference

## Overview

This project uses **envtest** from `controller-runtime` for integration testing, following the [Kubebuilder testing pattern](https://book.kubebuilder.io/cronjob-tutorial/writing-tests).

## Why envtest?

envtest is **perfect for your controller** because it provides:
- ✅ Real Kubernetes API server (not mocked)
- ✅ Full dynamic client and discovery API support
- ✅ Working informers with real watch events
- ✅ Fast startup (1-3 seconds vs 10-30 for kind)
- ✅ No Docker required
- ✅ CI/CD friendly

See [ENVTEST_EXPLAINED.md](./ENVTEST_EXPLAINED.md) for detailed explanation.

## Test Files

### 1. `internal/controller/suite_test.go`
Test suite setup that:
- Starts envtest API server and etcd
- Provides REST config (`cfg`) for all tests
- Sets up logging and test environment
- Cleans up after tests

### 2. `internal/controller/controller_envtest_test.go`
Integration tests that verify:
- Discovery API functionality
- Controller lifecycle (start/stop)
- Informer cache syncing
- Resource counting (create/delete)
- Namespace filtering

### 3. Unit Tests (Already Implemented)
- `internal/counter/counter_test.go` - Counter logic tests
- `internal/config/config_test.go` - Config validation tests
- `main_test.go` - Main helper function tests

## Running Tests

### All Tests (Unit + Integration)
```bash
make test
# or
ginkgo -v ./...
```

### Unit Tests Only (Fast)
```bash
make test-unit
# or
ginkgo -v -short ./...
```

### Integration Tests Only (envtest)
```bash
make test-integration
# or
ginkgo -v ./internal/controller/
# or
go test -v ./internal/controller/
```

### With Coverage
```bash
make test-coverage
```

## What envtest Provides

| Component | Provided? | Needed? |
|-----------|-----------|---------|
| API Server | ✅ Yes | ✅ Yes |
| etcd | ✅ Yes | ✅ Yes |
| Discovery API | ✅ Yes | ✅ Yes |
| Watch API | ✅ Yes | ✅ Yes |
| Dynamic Client | ✅ Yes | ✅ Yes |
| kubelet | ❌ No | ❌ No (not needed) |
| Pod Scheduling | ❌ No | ❌ No (not needed) |

**Result:** envtest provides exactly what you need!

## Key Concepts

### 1. Test Suite Setup
```go
// BeforeSuite - Runs once before all tests
var _ = BeforeSuite(func() {
    testEnv = &envtest.Environment{...}
    cfg, err := testEnv.Start()  // Starts API server
    // cfg is the REST config pointing to test API server
})
```

### 2. Using REST Config in Tests
```go
// In your tests, use cfg (REST config from suite_test.go)
controller, err := NewGVKController(controllerCfg, cfg, scheme)
//                                                       ^^^
//                                          This is the REST config from envtest
```

### 3. Real API Server Operations
```go
// Create resources - they go to real API server
deployment := &unstructured.Unstructured{...}
controller.dynamicClient.Resource(gvr).Create(ctx, deployment, ...)

// Controller informer receives real watch events!
Eventually(func() int {
    return controller.counter.GetCount(gvk)
}).Should(Equal(1))
```

## Example Test Flow

```
1. BeforeSuite starts envtest API server
   ↓
2. Test creates controller with cfg (REST config)
   ↓
3. Controller connects to envtest API server
   ↓
4. Test creates resource using dynamic client
   ↓
5. Real API server receives create request
   ↓
6. Controller informer receives watch event
   ↓
7. Event handler increments counter
   ↓
8. Test verifies counter was incremented
   ↓
9. AfterSuite stops API server
```

## Troubleshooting

### Tests are slow
- First run: envtest downloads API server binaries (one-time)
- Subsequent runs: Should be fast (1-3 seconds)

### Tests fail to start
- Check that controller-runtime is installed: `go mod tidy`
- Check logs in `BeforeSuite` for errors

### Resources not being counted
- Ensure informer cache has synced: `Eventually(func() { informer.HasSynced() })`
- Check that resources are created in correct namespace
- Verify namespace filtering if applicable

## Next Steps

1. ✅ Unit tests with Ginkgo/Gomega - **DONE**
2. ✅ Integration tests with envtest - **DONE**
3. ⬜ (Optional) E2E tests with kind if needed for full cluster testing

## References

- [Kubebuilder Testing Guide](https://book.kubebuilder.io/cronjob-tutorial/writing-tests)
- [envtest Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)
- [ENVTEST_EXPLAINED.md](./ENVTEST_EXPLAINED.md) - Detailed explanation
- [TESTING_STRATEGY.md](./TESTING_STRATEGY.md) - Complete testing strategy
