# Sample 1: Basic Reconciler

This example demonstrates the fundamental concepts of controller-runtime:

- Creating a Manager
- Implementing the `Reconcile()` function
- Defining a Custom Resource Definition (CRD)
- Status subresource management
- Request/Result pattern

## What You'll Learn

1. **Manager Setup**: How to create and configure a controller-runtime Manager
2. **Basic Reconciliation**: The core reconcile loop pattern
3. **CRD Definition**: Using kubebuilder markers for CRD generation
4. **Status Updates**: How to properly update status using the status subresource
5. **RBAC Generation**: Automatic RBAC rule generation from markers

## Prerequisites

- Go 1.24+
- kubectl configured with a Kubernetes cluster
- controller-gen: `go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest`

## Quick Start

### 1. Generate the CRD

```bash
# From the example directory
cd examples/01-basic-reconciler

# Generate DeepCopy methods and CRD manifests
controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="./..."
controller-gen crd paths="./..." output:crd:artifacts:config=config/crd
```

### 2. Install the CRD

```bash
kubectl apply -f config/crd/
```

Verify the CRD is installed:

```bash
kubectl get crds guestbooks.examples.k8s.io
```

### 3. Run the Controller

```bash
go run main.go
```

You should see output like:

```
INFO    setup   starting manager
INFO    controller-runtime.metrics      Starting metrics server
INFO    controller-runtime.health       Starting health check server
```

### 4. Create a Guestbook Resource

In another terminal:

```bash
kubectl apply -f config/samples/guestbook_sample.yaml
```

### 5. Watch the Controller Logs

You should see reconciliation logs:

```
INFO    Reconciling Guestbook   {"name": "guestbook-sample", "namespace": "default", ...}
INFO    Processing guestbook message    {"message": "Welcome to...", "replicas": 3}
INFO    Successfully reconciled Guestbook
```

### 6. Check the Resource Status

```bash
kubectl get guestbook guestbook-sample -o yaml
```

You should see the status updated:

```yaml
status:
  lastUpdated: "2025-01-XX..."
  observedGeneration: 1
  state: Active
```

Or use the custom columns:

```bash
kubectl get guestbooks
```

Output:
```
NAME               MESSAGE                              REPLICAS   STATE    AGE
guestbook-sample   Welcome to the Kubernetes Guestbook! 3          Active   1m
```

## Key Concepts Explained

### 1. The Reconcile Loop

```go
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the resource
    // 2. Process business logic
    // 3. Update status
    // 4. Return Result (requeue if needed)
}
```

The reconciler is called whenever:
- A Guestbook resource is created, updated, or deleted
- A watched resource changes
- A previous reconciliation returned `Result{Requeue: true}`

### 2. Request/Result Pattern

```go
// Success - don't requeue
return ctrl.Result{}, nil

// Success - requeue after duration
return ctrl.Result{RequeueAfter: time.Minute}, nil

// Success - requeue immediately
return ctrl.Result{Requeue: true}, nil

// Error - will requeue with exponential backoff
return ctrl.Result{}, err
```

### 3. Status Subresource

```go
// WRONG - updates both spec and status, triggers reconciliation
r.Update(ctx, guestbook)

// CORRECT - only updates status, doesn't trigger reconciliation
r.Status().Update(ctx, guestbook)
```

The status subresource prevents reconciliation loops when updating status.

### 4. Kubebuilder Markers

In `guestbook_types.go`:

```go
// +kubebuilder:object:root=true              // This is a root API object
// +kubebuilder:subresource:status            // Enable status subresource
// +kubebuilder:printcolumn:...               // Custom kubectl columns
// +kubebuilder:validation:Required           // Field validation
```

These markers generate:
- CRD YAML manifests
- OpenAPI validation schemas
- Custom kubectl output columns

### 5. RBAC Markers

In the controller:

```go
// +kubebuilder:rbac:groups=examples.k8s.io,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
```

These generate RBAC Role/ClusterRole manifests (when using kubebuilder scaffolding).

## Code Walkthrough

### main.go

1. **Initialize the Scheme**: Register your API types
2. **Create Manager**: Configure manager with options
3. **Setup Controller**: Register the reconciler
4. **Add Health Checks**: Liveness and readiness probes
5. **Start Manager**: Run until interrupted

### controller/guestbook_controller.go

1. **Fetch Resource**: Get the Guestbook from the API server
2. **Handle Not Found**: Gracefully handle deleted resources
3. **Process Logic**: Your business logic goes here
4. **Update Status**: Update the status subresource
5. **Return Result**: Determine if requeue is needed

### api/v1alpha1/guestbook_types.go

1. **Spec**: Desired state (what user wants)
2. **Status**: Observed state (what controller sees)
3. **Validation**: Kubebuilder markers for field validation
4. **Metadata**: Additional CRD configuration

## Testing

### Run Unit Tests

```bash
go test ./controller/... -v
```

### Run with envtest

```bash
# Install envtest binaries
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
setup-envtest use 1.31.0

# Run tests
go test ./controller/... -v
```

The tests use a real Kubernetes API server (via envtest) for integration testing.

## Experimentation Ideas

1. **Add Validation**: Modify the CRD to reject messages containing certain words
2. **Add Defaulting**: Set default values for optional fields
3. **Add Events**: Record Kubernetes events when processing
4. **Modify Status**: Add more status fields (e.g., `ReadyReplicas`)
5. **Add Finalizers**: Practice cleanup logic (covered in sample 03)
6. **Add Child Resources**: Create ConfigMaps based on Guestbook (covered in sample 04)

## Common Patterns

### Ignore Delete Events

```go
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    guestbook := &examplesv1alpha1.Guestbook{}
    err := r.Get(ctx, req.NamespacedName, guestbook)
    if err != nil {
        if errors.IsNotFound(err) {
            // Deleted - don't requeue
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }
    // ... continue processing
}
```

### Conditional Requeue

```go
if needsRetry {
    return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}
return ctrl.Result{}, nil
```

### Status Conditions (Preview)

While this example uses simple status fields, production controllers typically use "Conditions":

```go
// See sample 05-status-conditions for full implementation
status:
  conditions:
  - type: Ready
    status: "True"
    lastTransitionTime: "2025-01-15T..."
    reason: ReconcileSuccess
```

## Troubleshooting

### CRD Not Found

```bash
# Ensure CRD is installed
kubectl get crds | grep guestbook

# Reinstall if needed
kubectl apply -f config/crd/
```

### Controller Not Reconciling

1. Check controller logs for errors
2. Verify RBAC permissions (if running in-cluster)
3. Ensure resource is in a namespace the controller watches

### Status Not Updating

- Ensure the CRD has `+kubebuilder:subresource:status` marker
- Regenerate CRDs after marker changes
- Use `r.Status().Update()` not `r.Update()`

## Next Steps

- **Sample 02**: Learn about predicates and event filtering
- **Sample 03**: Implement finalizers for cleanup logic
- **Sample 05**: Use proper status conditions

## Additional Resources

- [Controller-Runtime Godoc](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller-Runtime FAQ](https://github.com/kubernetes-sigs/controller-runtime/blob/main/FAQ.md)
