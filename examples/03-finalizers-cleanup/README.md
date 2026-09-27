# Sample 3: Finalizers & Cleanup 🔥 PRIORITY

This example demonstrates how to use finalizers to perform cleanup logic when a resource is deleted. This is essential when your controller manages external resources that need explicit cleanup.

## What You'll Learn

1. **Finalizer Registration**: How to add finalizers using `controllerutil.AddFinalizer()`
2. **Deletion Detection**: Checking the `DeletionTimestamp` to detect when a resource is being deleted
3. **Cleanup Logic**: Performing cleanup of external resources before allowing deletion
4. **Finalizer Removal**: Removing finalizers with `controllerutil.RemoveFinalizer()`
5. **Safe Deletion**: Preventing orphaned external resources

## Use Cases

Use finalizers when your controller:
- Manages external resources (databases, cloud resources, DNS records)
- Needs to clean up child resources in a specific order
- Must notify external systems before deletion
- Needs to perform graceful shutdown or deprovisioning

## Prerequisites

- Go 1.24+
- kubectl configured with a Kubernetes cluster
- controller-gen: `go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest`

## Quick Start

### 1. Generate and Install CRD

```bash
cd examples/03-finalizers-cleanup

# Generate code and CRDs
controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="./..."
controller-gen crd paths="./..." output:crd:artifacts:config=config/crd

# Install CRD
kubectl apply -f config/crd/
```

### 2. Run the Controller

```bash
go run main.go
```

### 3. Create a Database

```bash
kubectl apply -f config/samples/database_sample.yaml
```

### 4. Observe the Finalizer

```bash
kubectl get database database-sample -o yaml
```

You'll see the finalizer in the metadata:

```yaml
metadata:
  finalizers:
  - finalizers.examples.k8s.io/database-cleanup
```

And the status:

```yaml
status:
  externalID: db-default-database-sample
  message: Database is ready
  observedGeneration: 1
  state: Ready
```

### 5. Delete the Database

```bash
kubectl delete database database-sample
```

Watch the controller logs - you'll see:

```
INFO    Database is being deleted, performing cleanup
INFO    Cleaning up external database   {"externalID": "db-default-database-sample"}
INFO    External database deleted successfully
INFO    Finalizer removed, Database will be deleted
```

The resource won't be deleted until the cleanup completes!

## Key Concepts

### 1. The Finalizer Pattern

```go
const databaseFinalizer = "finalizers.examples.k8s.io/database-cleanup"

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Check if being deleted
    if database.ObjectMeta.DeletionTimestamp.IsZero() {
        // Not being deleted - ensure finalizer exists
        if !controllerutil.ContainsFinalizer(database, databaseFinalizer) {
            controllerutil.AddFinalizer(database, databaseFinalizer)
            r.Update(ctx, database)
        }
    } else {
        // Being deleted - perform cleanup
        if controllerutil.ContainsFinalizer(database, databaseFinalizer) {
            // Do cleanup work
            r.cleanupExternalResources(ctx, database)

            // Remove finalizer to allow deletion
            controllerutil.RemoveFinalizer(database, databaseFinalizer)
            r.Update(ctx, database)
        }
        return ctrl.Result{}, nil
    }

    // Normal reconciliation
    // ...
}
```

### 2. DeletionTimestamp

When a resource is deleted:
- Kubernetes sets the `DeletionTimestamp` field
- The resource enters a "terminating" state
- If finalizers exist, the resource is NOT deleted yet
- Once all finalizers are removed, Kubernetes deletes the resource

```go
if !database.ObjectMeta.DeletionTimestamp.IsZero() {
    // Resource is being deleted
}
```

### 3. Finalizer Names

Use a unique, descriptive finalizer name:

```go
// GOOD - unique and descriptive
const databaseFinalizer = "finalizers.examples.k8s.io/database-cleanup"

// BAD - too generic, could conflict
const finalizer = "cleanup"
```

### 4. Cleanup Idempotency

Cleanup logic must be idempotent (safe to run multiple times):

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, database *Database) error {
    externalID := database.Status.ExternalID
    if externalID == "" {
        // Nothing to cleanup
        return nil
    }

    // Check if resource still exists before deleting
    if !r.externalClient.Exists(externalID) {
        return nil  // Already deleted
    }

    // Perform deletion
    return r.externalClient.Delete(externalID)
}
```

### 5. Handling Cleanup Failures

If cleanup fails, the controller should:
1. Log the error
2. Return the error to trigger a retry
3. NOT remove the finalizer

```go
if err := r.cleanupExternalResources(ctx, database); err != nil {
    logger.Error(err, "Failed to cleanup external resources")
    return ctrl.Result{}, err  // Will retry with backoff
}
```

## Testing Finalizers

### Test Deletion Protection

```bash
# Create a database
kubectl apply -f config/samples/database_sample.yaml

# Stop the controller (Ctrl+C)

# Try to delete the database
kubectl delete database database-sample

# The database will hang in "Terminating" state
kubectl get databases
NAME               DATABASE      STATE   EXTERNAL ID   AGE
database-sample    myapp-prod    Ready   db-...        1m    (Terminating)

# Restart the controller - cleanup will complete and resource will be deleted
```

### Test Cleanup Logic

```bash
# Create and delete rapidly
kubectl apply -f config/samples/database_sample.yaml
sleep 2
kubectl delete database database-sample

# Watch logs to ensure cleanup runs
```

## Experimentation Ideas

1. **Simulate Cleanup Failure**: Make `cleanupExternalResources` return an error and observe retry behavior

2. **Add Delay**: Add a sleep in cleanup to see the resource stay in Terminating state longer

3. **Multiple Finalizers**: Add a second finalizer to see how Kubernetes waits for all finalizers

4. **Cleanup Order**: Create resources with dependencies and use finalizers to ensure proper cleanup order

5. **Status During Deletion**: Update status to "Deleting" to inform users

## Common Patterns

### Pattern 1: Conditional Cleanup

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *Database) error {
    // Only cleanup if we actually created something
    if db.Status.ExternalID == "" {
        return nil
    }

    // Cleanup logic
    return r.deleteExternal(db.Status.ExternalID)
}
```

### Pattern 2: Parallel Cleanup

```go
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, db *Database) error {
    var eg errgroup.Group

    eg.Go(func() error {
        return r.deleteDatabase(db.Status.ExternalID)
    })

    eg.Go(func() error {
        return r.deleteBackups(db.Status.ExternalID)
    })

    return eg.Wait()
}
```

### Pattern 3: Status Update During Deletion

```go
if database.ObjectMeta.DeletionTimestamp != nil {
    // Update status to show we're deleting
    database.Status.State = "Deleting"
    database.Status.Message = "Cleaning up external resources"
    r.Status().Update(ctx, database)

    // Perform cleanup
    // ...
}
```

## Troubleshooting

### Resource Stuck in Terminating State

**Cause**: Finalizer not removed, or controller not running

**Solution**:
```bash
# Check if controller is running
kubectl get pods -n controller-namespace

# Manually remove finalizer (use with caution!)
kubectl patch database database-sample -p '{"metadata":{"finalizers":[]}}' --type=merge
```

### Cleanup Not Running

**Cause**: Finalizer not registered before deletion

**Solution**: Ensure finalizer is added during creation, not just when needed

### External Resources Orphaned

**Cause**: Controller crashed during cleanup, or cleanup logic has bugs

**Solution**:
- Make cleanup idempotent
- Add retry logic
- Keep track of external IDs in status

## Best Practices

1. **Add Finalizers Early**: Add finalizers during resource creation, not deletion
2. **Use Unique Names**: Namespace your finalizer names to avoid conflicts
3. **Idempotent Cleanup**: Ensure cleanup can be safely run multiple times
4. **Handle Failures**: Return errors to trigger retries, don't silently fail
5. **Update Status**: Show deletion progress in status
6. **Timeout Protection**: Add timeouts to prevent hanging forever
7. **Multiple Finalizers**: Use separate finalizers for independent cleanup tasks

## Security Considerations

1. **RBAC**: Ensure controller has permission to update finalizers:
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: database-controller
   rules:
   - apiGroups: ["finalizers.examples.k8s.io"]
     resources: ["databases/finalizers"]
     verbs: ["update"]
   ```

2. **Prevent Abuse**: Users could add finalizers to prevent deletion - implement admission webhooks if needed

## Next Steps

- **Sample 04**: Learn about owner references for automatic garbage collection
- **Sample 05**: Implement proper status conditions
- **Sample 07**: Use webhooks to validate finalizer requirements

## Additional Resources

- [Kubernetes Finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
- [Using Finalizers](https://book.kubebuilder.io/reference/using-finalizers.html)
- [Controller-Runtime Finalizer Utils](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#AddFinalizer)
