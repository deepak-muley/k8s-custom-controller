# Controller-Runtime Guide

This guide helps you understand when and how to use controller-runtime vs client-go for Kubernetes controller development.

## Table of Contents

- [Overview](#overview)
- [Quick Decision Matrix](#quick-decision-matrix)
- [When to Use Controller-Runtime](#when-to-use-controller-runtime)
- [When to Use Client-Go](#when-to-use-client-go)
- [Feature Comparison](#feature-comparison)
- [All Examples Overview](#all-examples-overview)
- [Learning Path](#learning-path)
- [Migration Considerations](#migration-considerations)

## Overview

This project demonstrates **both approaches**:
- **Main Controller**: Built with `client-go` (direct informers + workqueue)
- **Examples**: Built with `controller-runtime` (manager + reconcile pattern)

Both are valid, production-ready approaches. The choice depends on your specific requirements.

## Quick Decision Matrix

| Scenario | Recommended Approach | Why |
|----------|---------------------|-----|
| New CRD controller | **controller-runtime** | Less boilerplate, built-in best practices |
| Extending existing client-go project | **client-go** | Consistency with existing code |
| Simple controller (1-2 resources) | Either (slight edge to **controller-runtime**) | Both work well, controller-runtime slightly simpler |
| Complex multi-resource controller | **controller-runtime** | Better abstractions for watches and ownership |
| Need fine-grained caching control | **client-go** | Direct informer access |
| Webhooks required | **controller-runtime** | Built-in webhook server |
| Learning Kubernetes controllers | Start with **controller-runtime** | Clearer patterns, less complexity |
| Building low-level k8s components | **client-go** | More control, used by k8s core |

## When to Use Controller-Runtime

✅ **Choose controller-runtime when:**

### 1. Building CRD Controllers

Controller-runtime excels at CRD-based controllers:

```go
// Simple CRD controller with controller-runtime
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fetch resource
    resource := &myapi.MyResource{}
    if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Business logic
    // ...

    return ctrl.Result{}, nil
}
```

**Benefits:**
- Automatic reconciliation loop
- Built-in error handling and retry logic
- Status subresource support
- Webhook integration

### 2. You Need Webhooks

Controller-runtime provides a complete webhook server:

```go
// Validation webhook - just implement the interface
func (r *MyResource) ValidateCreate() error {
    // Validation logic
    return nil
}

// Mutating webhook
func (r *MyResource) Default() {
    // Set defaults
}
```

See [examples/07-webhooks](../examples/07-webhooks) for details.

### 3. Multiple Resource Types

Controller-runtime makes watching multiple resources elegant:

```go
ctrl.NewControllerManagedBy(mgr).
    For(&appsv1.Deployment{}).       // Primary resource
    Owns(&corev1.Pod{}).              // Watch owned pods
    Owns(&corev1.Service{}).          // Watch owned services
    Complete(r)
```

See [examples/06-multi-resource-watch](../examples/06-multi-resource-watch).

### 4. Standard Production Features

Controller-runtime includes:
- **Leader Election**: Built-in, just set a flag
- **Metrics**: Prometheus metrics out of the box
- **Health Checks**: `/healthz` and `/readyz` endpoints
- **Caching**: Smart caching with indexing support

### 5. Team Productivity

Controller-runtime reduces boilerplate:
- ~200 lines for a basic controller-runtime controller
- ~500 lines for equivalent client-go controller
- Less error-prone
- Easier to onboard new developers

## When to Use Client-Go

✅ **Choose client-go when:**

### 1. Fine-Grained Control Needed

Direct informer control for specific use cases:

```go
// client-go: Custom informer configuration
informer := cache.NewSharedIndexInformer(
    listWatcher,
    &corev1.Pod{},
    resyncPeriod,
    cache.Indexers{
        "customIndex": myCustomIndexFunc,
    },
)

// Full control over event handlers
informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
        // Custom add logic
    },
    UpdateFunc: func(old, new interface{}) {
        // Custom update logic with old/new comparison
    },
    DeleteFunc: func(obj interface{}) {
        // Custom delete logic
    },
})
```

**Benefits:**
- Access to `old` and `new` objects in UpdateFunc
- Custom resync periods per informer
- Direct workqueue manipulation

### 2. Non-CRD Resources

For built-in resources or unusual patterns:

```go
// Watching namespaces with custom logic
nsInformer := cache.NewSharedIndexInformer(
    &cache.ListWatch{
        ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
            return clientset.CoreV1().Namespaces().List(ctx, options)
        },
        WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
            return clientset.CoreV1().Namespaces().Watch(ctx, options)
        },
    },
    &corev1.Namespace{},
    0,
    cache.Indexers{},
)
```

### 3. Extending Existing Codebase

If your project already uses client-go extensively, stay consistent.

### 4. Learning Kubernetes Internals

Client-go teaches you how Kubernetes controllers actually work:
- Informer pattern
- Workqueue behavior
- Caching mechanics
- Event processing

See the main controller in this project for a complete example.

### 5. Performance Optimization

When you need to optimize:
- Custom batch processing
- Specific cache eviction strategies
- Manual rate limiting configuration
- Custom metrics and profiling

## Feature Comparison

| Feature | controller-runtime | client-go | Example |
|---------|-------------------|-----------|---------|
| **Reconciliation** | Automatic `Reconcile()` loop | Manual workqueue + handlers | 01-basic-reconciler |
| **CRD Development** | Kubebuilder markers + codegen | Manual YAML | 01-basic-reconciler |
| **Status Updates** | `Status().Update()` | Manual subresource calls | 05-status-conditions |
| **Finalizers** | `controllerutil` helpers | Manual string manipulation | 03-finalizers-cleanup |
| **Owner References** | `SetControllerReference()` | Manual field setting | 04-owner-references |
| **Webhooks** | Built-in server | Custom server setup | 07-webhooks |
| **Metrics** | Automatic Prometheus | Manual instrumentation | 08-metrics-events |
| **Leader Election** | Single flag | Manual lock implementation | N/A |
| **Testing** | `envtest` (real API server) | Mock clients or integration | All examples |
| **Event Filtering** | `predicate.Funcs` | Custom filter logic | 02-predicates-filtering |
| **Multi-Resource Watch** | `Owns()`, `Watches()` | Multiple informers | 06-multi-resource-watch |
| **Indexing** | `GetFieldIndexer()` | `cache.Indexers` | 09-advanced-indexing |
| **Rate Limiting** | Workqueue options | Manual workqueue config | 11-rate-limiting-backoff |
| **External Events** | `source.Channel` | Custom event injection | 10-event-source-chaining |
| **Learning Curve** | Moderate | Steeper | N/A |
| **Boilerplate** | Low | Higher | N/A |
| **Control** | High-level | Low-level | N/A |
| **Flexibility** | Good | Excellent | N/A |

## All Examples Overview

### Foundation (Start Here)

#### [01-basic-reconciler](../examples/01-basic-reconciler)
**Concepts:** Manager setup, Reconcile loop, CRD definition, Status updates

The starting point for all controller-runtime projects. Learn the fundamental `Reconcile()` pattern.

#### [02-predicates-filtering](../examples/02-predicates-filtering)
**Concepts:** Event filtering, Performance optimization

Reduce unnecessary reconciliations by filtering events with predicates.

#### [03-finalizers-cleanup](../examples/03-finalizers-cleanup) 🔥
**Concepts:** Cleanup logic, External resource management

Essential for controllers that manage external resources (databases, cloud resources, etc.).

### Advanced Patterns

#### [04-owner-references](../examples/04-owner-references)
**Concepts:** Resource ownership, Garbage collection

Automatically clean up child resources when parent is deleted.

#### [05-status-conditions](../examples/05-status-conditions) 🔥
**Concepts:** Status management, Conditions

Production pattern for exposing controller state to users and other controllers.

#### [06-multi-resource-watch](../examples/06-multi-resource-watch)
**Concepts:** Multiple resource types, Cross-resource reconciliation

React to changes in multiple resource types with elegant watch patterns.

### Production Features

#### [07-webhooks](../examples/07-webhooks) 🔥
**Concepts:** Validation webhooks, Mutating webhooks, Admission control

Add validation and defaulting logic to your CRDs.

#### [08-metrics-events](../examples/08-metrics-events)
**Concepts:** Observability, Prometheus metrics, Events

Essential for production: expose metrics and record events.

#### [09-advanced-indexing](../examples/09-advanced-indexing)
**Concepts:** Cache optimization, Fast lookups

Optimize performance with custom cache indexes.

#### [10-event-source-chaining](../examples/10-event-source-chaining) ⭐
**Concepts:** External events, Non-Kubernetes triggers

Integrate with external systems (webhooks, alerts, monitoring).

#### [11-rate-limiting-backoff](../examples/11-rate-limiting-backoff) ⭐
**Concepts:** Rate limiting, Retry strategies

Control reconciliation rate and implement smart retry logic.

## Learning Path

### Beginner Path (New to Controllers)

1. **Read the main controller code** (`main.go`, `internal/controller/`) - Understand client-go patterns
2. **Example 01**: Basic Reconciler - See controller-runtime equivalent
3. **Example 03**: Finalizers - Learn cleanup patterns
4. **Example 05**: Status Conditions - Production status management
5. **Build your first CRD controller**

### Intermediate Path (Know Basics)

1. **Example 02**: Predicates - Optimize performance
2. **Example 04**: Owner References - Resource relationships
3. **Example 06**: Multi-Resource Watch - Complex watches
4. **Example 07**: Webhooks - Admission control
5. **Example 08**: Metrics & Events - Observability

### Advanced Path (Production-Ready)

1. **Example 09**: Advanced Indexing - Cache optimization
2. **Example 10**: Event Source Chaining - External integration
3. **Example 11**: Rate Limiting - Control reconciliation
4. **Read [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)** - Convert client-go to controller-runtime
5. **Implement HA with leader election**

## Migration Considerations

### From client-go to controller-runtime

**Pros:**
- ✅ Less code to maintain
- ✅ Built-in best practices
- ✅ Easier webhook integration
- ✅ Better testing with envtest

**Cons:**
- ⚠️ Learning curve for the team
- ⚠️ Less control over informer details
- ⚠️ May need refactoring

See [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) for step-by-step migration.

### From controller-runtime to client-go

**Pros:**
- ✅ More fine-grained control
- ✅ Better for unusual patterns
- ✅ Can optimize specific behaviors

**Cons:**
- ⚠️ More boilerplate code
- ⚠️ Manual implementation of common patterns
- ⚠️ More testing complexity

**Rarely recommended** - only for specific performance or control needs.

## Best Practices

### For controller-runtime Projects

1. **Use Predicates**: Filter events early (Example 02)
2. **Implement Finalizers**: Clean up external resources (Example 03)
3. **Use Status Conditions**: Standard status pattern (Example 05)
4. **Add Webhooks**: Validate before admission (Example 07)
5. **Expose Metrics**: Essential for production (Example 08)
6. **Index Frequently**: Optimize cache lookups (Example 09)
7. **Test with envtest**: Real API server testing (All examples)

### For client-go Projects

1. **Use SharedInformers**: Don't create duplicate watchers
2. **Implement Proper Synchronization**: Relist periodically
3. **Handle Deletions**: Check for `DeletedFinalStateUnknown`
4. **Use Workqueue**: Don't process in event handlers
5. **Add Indexes**: Optimize cache lookups
6. **Test Thoroughly**: Mock or integration tests

## Common Pitfalls

### controller-runtime

❌ **Updating Status in Reconcile Body**
```go
// WRONG
r.Update(ctx, resource)  // Triggers reconcile

// RIGHT
r.Status().Update(ctx, resource)  // No reconcile
```

❌ **Forgetting to Handle Deletion**
```go
// WRONG - crashes on deleted resources
resource := &MyResource{}
r.Get(ctx, req.NamespacedName, resource)

// RIGHT
if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
}
```

❌ **Not Using Finalizers**
```go
// If managing external resources, MUST use finalizers
// See Example 03
```

### client-go

❌ **Processing in Event Handlers**
```go
// WRONG - blocks informer
AddFunc: func(obj interface{}) {
    process(obj)  // Expensive operation
}

// RIGHT - queue for later
AddFunc: func(obj interface{}) {
    key, _ := cache.MetaNamespaceKeyFunc(obj)
    queue.Add(key)
}
```

❌ **Not Checking DeletedFinalStateUnknown**
```go
// WRONG
pod := obj.(*corev1.Pod)

// RIGHT
pod, ok := obj.(*corev1.Pod)
if !ok {
    tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
    if !ok {
        return fmt.Errorf("unexpected object type")
    }
    pod, ok = tombstone.Obj.(*corev1.Pod)
}
```

## Performance Considerations

### controller-runtime

- **Caching**: Automatic caching per GVK
- **Indexing**: Add indexes for frequent queries (Example 09)
- **Predicates**: Filter events early (Example 02)
- **Rate Limiting**: Configure workqueue (Example 11)

### client-go

- **Shared Informers**: Reuse informers across controllers
- **Resync Period**: Balance freshness vs load
- **Custom Indexes**: Add for specific lookup patterns
- **Workqueue Config**: Tune rate limits and retries

## Resources

### controller-runtime
- [Official Docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Examples in this repo](../examples)

### client-go
- [Official Docs](https://github.com/kubernetes/client-go)
- [Sample Controller](https://github.com/kubernetes/sample-controller)
- [Main controller in this repo](../main.go)

### General Kubernetes
- [API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Controller Pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
- [CRD Documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)

## Conclusion

Both controller-runtime and client-go are excellent choices:

- **controller-runtime**: Higher-level, less boilerplate, modern patterns
- **client-go**: Lower-level, more control, used by Kubernetes core

**For most new controllers, we recommend controller-runtime.** It embeds years of best practices and handles common patterns automatically.

This project provides both approaches so you can learn from each and choose the right tool for your needs.

## Next Steps

1. **Explore the examples**: Start with [01-basic-reconciler](../examples/01-basic-reconciler)
2. **Compare with main controller**: See the client-go equivalent
3. **Read the migration guide**: [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)
4. **Build your own controller**: Apply what you've learned

Questions or feedback? Open an issue in the repository!
