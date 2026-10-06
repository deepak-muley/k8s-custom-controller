# Example 06: Multi-Resource Watch

This example demonstrates how to watch multiple resource types in a Kubernetes controller using controller-runtime. The DeploymentMonitor controller watches both Deployments and their ReplicaSets to provide comprehensive monitoring information.

## Overview

The DeploymentMonitor CRD allows you to monitor a Deployment and automatically tracks:
- Deployment status (desired, ready, and available replicas)
- All ReplicaSets owned by the Deployment
- ReplicaSet details including replicas, ready replicas, and revision information

## Key Concepts

### Multi-Resource Watching

This example demonstrates two patterns for watching multiple resources:

1. **Watches() with custom handler**: Used for Deployments and ReplicaSets
   - Uses `handler.EnqueueRequestsFromMapFunc` to map watched resources to our CRD
   - Allows watching resources that are not directly owned by our CRD
   - Provides fine-grained control over which reconciliation requests are triggered

2. **Custom event mapping functions**:
   - `findDeploymentMonitorsForDeployment`: Maps Deployment changes to DeploymentMonitor reconciliations
   - `findDeploymentMonitorsForReplicaSet`: Maps ReplicaSet changes through their owner Deployment to DeploymentMonitor reconciliations

### Why Watch Multiple Resources?

- **Comprehensive monitoring**: Get updates when either the Deployment or its ReplicaSets change
- **Automatic updates**: Status reflects the current state without manual polling
- **Efficient**: Controller-runtime's event system ensures minimal API calls
- **Real-time**: React immediately to changes in the cluster

## Files

- `api/v1alpha1/groupversion_info.go` - API group definition (watches.examples.k8s.io)
- `api/v1alpha1/deploymentmonitor_types.go` - DeploymentMonitor CRD definition
- `controller/deploymentmonitor_controller.go` - Controller with multi-resource watching
- `main.go` - Main entry point
- `config/crd/` - Generated CRD manifests
- `config/samples/` - Sample DeploymentMonitor resources

## Building

From the repository root:

```bash
# Generate DeepCopy and CRD manifests
make generate manifests EXAMPLE=06-multi-resource-watch

# Or manually:
controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./examples/06-multi-resource-watch/api/v1alpha1"
controller-gen crd:crdVersions=v1 paths="./examples/06-multi-resource-watch/api/v1alpha1" output:crd:dir="./examples/06-multi-resource-watch/config/crd"

# Build the controller
go build -o bin/06-multi-resource-watch ./examples/06-multi-resource-watch
```

## Running

### Prerequisites

- A running Kubernetes cluster (kind, minikube, or other)
- kubectl configured to access the cluster

### Steps

1. Install the CRD:
```bash
kubectl apply -f examples/06-multi-resource-watch/config/crd/watches.examples.k8s.io_deploymentmonitors.yaml
```

2. Create a sample Deployment to monitor:
```bash
kubectl create deployment nginx-deployment --image=nginx:latest --replicas=3
```

3. Run the controller:
```bash
./bin/06-multi-resource-watch
```

4. In another terminal, create a DeploymentMonitor:
```bash
kubectl apply -f examples/06-multi-resource-watch/config/samples/deploymentmonitor_sample.yaml
```

5. Check the status:
```bash
kubectl get deploymentmonitors
kubectl get deploymentmonitor nginx-monitor -o yaml
```

## Example Output

```bash
$ kubectl get deploymentmonitors
NAME            DEPLOYMENT          FOUND   DESIRED   AVAILABLE   REPLICASETS   AGE
nginx-monitor   nginx-deployment    true    3         3           1             2m

$ kubectl get deploymentmonitor nginx-monitor -o yaml
apiVersion: watches.examples.k8s.io/v1alpha1
kind: DeploymentMonitor
metadata:
  name: nginx-monitor
  namespace: default
spec:
  deploymentName: nginx-deployment
  namespace: default
status:
  availableReplicas: 3
  deploymentFound: true
  desiredReplicas: 3
  lastUpdated: "2025-10-05T16:37:00Z"
  message: "Deployment nginx-deployment is ready with 3/3 replicas available"
  observedGeneration: 1
  readyReplicas: 3
  replicaSets:
  - availableReplicas: 3
    name: nginx-deployment-7c5ddbdf54
    readyReplicas: 3
    replicas: 3
    revision: "1"
  totalReplicaSets: 1
```

## Testing the Multi-Resource Watch

1. **Scale the Deployment**:
```bash
kubectl scale deployment nginx-deployment --replicas=5
```
Watch the controller logs - it will reconcile due to the Deployment change.

2. **Trigger a rolling update**:
```bash
kubectl set image deployment/nginx-deployment nginx=nginx:1.21
```
Watch the controller logs - it will reconcile as new ReplicaSets are created and old ones are scaled down.

3. **Delete the Deployment**:
```bash
kubectl delete deployment nginx-deployment
```
The DeploymentMonitor status will update to show the deployment is not found.

## Controller Logic

### SetupWithManager

```go
func (r *DeploymentMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&watchesv1alpha1.DeploymentMonitor{}).
        Watches(
            &appsv1.Deployment{},
            handler.EnqueueRequestsFromMapFunc(r.findDeploymentMonitorsForDeployment),
        ).
        Watches(
            &appsv1.ReplicaSet{},
            handler.EnqueueRequestsFromMapFunc(r.findDeploymentMonitorsForReplicaSet),
        ).
        Named("deploymentmonitor").
        Complete(r)
}
```

- `For()` - Primary watch on DeploymentMonitor resources
- `Watches()` - Secondary watches on Deployments and ReplicaSets
- `handler.EnqueueRequestsFromMapFunc` - Maps events to reconciliation requests

### Event Mapping Functions

The controller implements two mapping functions:

1. **findDeploymentMonitorsForDeployment**: When a Deployment changes, finds all DeploymentMonitors that reference it
2. **findDeploymentMonitorsForReplicaSet**: When a ReplicaSet changes, finds the owning Deployment, then finds all DeploymentMonitors that reference that Deployment

## RBAC Permissions

The controller requires the following permissions:

```yaml
# DeploymentMonitor CRD
- apiGroups: ["watches.examples.k8s.io"]
  resources: ["deploymentmonitors", "deploymentmonitors/status"]
  verbs: ["get", "list", "watch", "update", "patch"]

# Deployments (read-only)
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch"]

# ReplicaSets (read-only)
- apiGroups: ["apps"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]
```

## Use Cases

1. **Deployment Monitoring Dashboard**: Track the health of multiple deployments
2. **Rolling Update Observer**: Monitor the progress of rolling updates
3. **Replica Set History**: Keep track of all ReplicaSets and their revisions
4. **Alert System**: Trigger alerts when deployments are not healthy
5. **Custom Metrics**: Export deployment health metrics to monitoring systems

## Key Takeaways

1. **Multiple watches**: Use `Watches()` to watch resources beyond the primary resource
2. **Event handlers**: Use `handler.EnqueueRequestsFromMapFunc` for custom event routing
3. **Mapping functions**: Implement functions to map watched resources to reconciliation requests
4. **Owner references**: Use owner references to find relationships between resources
5. **Status updates**: Keep status up-to-date based on multiple resource states

## Comparison with Owns()

This example uses `Watches()` instead of `Owns()` because:
- The Deployments and ReplicaSets are NOT owned by our DeploymentMonitor CRD
- We're monitoring existing resources, not creating/managing them
- `Owns()` would only work for resources we create with owner references pointing to our CRD
- `Watches()` with custom handlers gives us more flexibility

## Next Steps

- Example 07: Webhooks - Add validation and mutation webhooks
- Example 08: Metrics & Events - Add metrics and Kubernetes events
- Example 09: Advanced Indexing - Use field indexing for efficient queries

## References

- [Controller Runtime Watches](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/builder#Builder.Watches)
- [Event Handlers](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/handler)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes ReplicaSets](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/)
