# Controller-Runtime Examples

This directory contains comprehensive examples demonstrating controller-runtime features and patterns. These examples complement the existing client-go based controller in the main project and showcase modern Kubernetes controller development.

## Overview

These 11 samples progress from basic concepts to advanced production patterns, providing a complete learning path for controller-runtime development.

## Prerequisites

- Go 1.24+
- kubectl configured with a Kubernetes cluster (local or remote)
- controller-gen tool: `go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest`
- Optional: cert-manager (for webhooks example)

## Quick Start

```bash
# Generate CRDs for all examples
make examples-generate

# Install CRDs to your cluster
make examples-install-crds

# Run a specific example
make example-run EXAMPLE=01-basic-reconciler

# Run tests for all examples
make examples-test

# Build all example binaries
make examples-build
```

## Learning Path

### Phase 1: Foundation (Start Here)

#### 1. [Basic Reconciler](01-basic-reconciler/)
**What it demonstrates:**
- Manager setup with `ctrl.NewManager()`
- Basic `Reconcile()` implementation
- CRD definition with kubebuilder markers
- Get/Update operations
- Request/Result pattern

**When to use:** Every controller-runtime project starts here.

#### 2. [Predicates & Filtering](02-predicates-filtering/)
**What it demonstrates:**
- Event filtering with `predicate.Funcs`
- Label/annotation selectors
- Generation change detection
- Reducing unnecessary reconciliations

**When to use:** When you need to optimize controller performance by filtering events.

#### 3. [Finalizers & Cleanup](03-finalizers-cleanup/) 🔥 PRIORITY
**What it demonstrates:**
- Finalizer registration with `controllerutil.AddFinalizer()`
- Cleanup logic on resource deletion
- External resource cleanup patterns
- Proper deletion handling

**When to use:** When your controller manages external resources (databases, cloud resources, etc.) that need cleanup.

### Phase 2: Advanced Patterns

#### 4. [Owner References](04-owner-references/)
**What it demonstrates:**
- `controllerutil.SetControllerReference()`
- Automatic garbage collection
- Parent-child resource relationships
- Cascading deletion

**When to use:** When your controller creates child resources that should be cleaned up with the parent.

#### 5. [Status Conditions](05-status-conditions/) 🔥 PRIORITY
**What it demonstrates:**
- Status subresource updates
- Condition management (Ready, Available, Progressing)
- `Status().Update()` vs `Update()`
- Standard condition types

**When to use:** Essential for exposing controller state to users and other controllers.

#### 6. [Multi-Resource Watch](06-multi-resource-watch/)
**What it demonstrates:**
- Watching multiple resource types
- Owner-based watches with `Owns()`
- External watches with `EnqueueRequestsFromMapFunc`
- Cross-resource reconciliation

**When to use:** When your controller needs to react to changes in multiple resource types.

### Phase 3: Production Features

#### 7. [Webhooks](07-webhooks/) 🔥 PRIORITY
**What it demonstrates:**
- Validation webhooks (`ValidateCreate/Update/Delete`)
- Mutating webhooks (`Default()`)
- Webhook server setup
- Certificate management with cert-manager

**When to use:** For resource validation, defaulting values, or enforcing policies.

#### 8. [Metrics & Events](08-metrics-events/)
**What it demonstrates:**
- Custom Prometheus metrics
- Event recording with `recorder.Event()`
- Health/readiness probes
- `/metrics` endpoint setup

**When to use:** Essential for production observability and debugging.

#### 9. [Advanced Indexing](09-advanced-indexing/)
**What it demonstrates:**
- Custom field indexing with `GetFieldIndexer().IndexField()`
- Fast cache lookups
- Performance optimization with `MatchingFields`
- Query optimization

**When to use:** When you need efficient lookups by custom fields (e.g., finding all resources with a specific label).

#### 10. [Event Source Chaining](10-event-source-chaining/) ⭐ NEW
**What it demonstrates:**
- External event sources with `source.Channel`
- Chaining events from non-Kubernetes sources
- Custom event handlers
- Integration with external systems (webhooks, alerts, etc.)

**When to use:** When your controller needs to react to events outside Kubernetes (GitHub webhooks, monitoring alerts, etc.).

#### 11. [Rate Limiting & Backoff](11-rate-limiting-backoff/) ⭐ NEW
**What it demonstrates:**
- Rate limiting with workqueue options
- Exponential backoff strategies
- `RateLimitingInterface` configuration
- Per-item and global rate limits

**When to use:** To prevent controller overload and implement smart retry logic.

## Sample Structure

Each sample follows a consistent structure:

```
XX-sample-name/
├── main.go                          # Manager setup + controller registration
├── api/v1alpha1/
│   ├── types.go                     # CRD definition with kubebuilder markers
│   └── zz_generated.deepcopy.go     # Generated by controller-gen
├── controller/
│   ├── controller.go                # Reconcile logic
│   ├── controller_test.go           # Unit tests
│   └── suite_test.go                # envtest integration tests
├── config/
│   ├── crd/                         # Generated CRD manifests
│   └── samples/                     # Example CR YAMLs
└── README.md                        # Sample-specific documentation
```

## Running Examples

### Option 1: In-Cluster (Recommended)

```bash
# Install CRDs
kubectl apply -f examples/01-basic-reconciler/config/crd/

# Run the controller
cd examples/01-basic-reconciler
go run main.go

# In another terminal, apply sample resources
kubectl apply -f config/samples/
```

### Option 2: Out-of-Cluster (Development)

```bash
# Set KUBECONFIG if needed
export KUBECONFIG=~/.kube/config

# Run the controller
cd examples/01-basic-reconciler
go run main.go --kubeconfig=$KUBECONFIG
```

### Option 3: Build and Deploy

```bash
# Build binary
make examples-build

# Run binary
./bin/01-basic-reconciler
```

## Testing

Each sample includes comprehensive tests:

```bash
# Run all example tests
make examples-test

# Run tests for a specific example
cd examples/01-basic-reconciler
go test ./... -v

# Run with coverage
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## CRD Development Workflow

```bash
# 1. Define types in api/v1alpha1/types.go with kubebuilder markers

# 2. Generate DeepCopy methods and CRD manifests
make examples-generate

# 3. Install CRDs to cluster
kubectl apply -f examples/XX-sample-name/config/crd/

# 4. Run controller
make example-run EXAMPLE=XX-sample-name

# 5. Test with sample resources
kubectl apply -f examples/XX-sample-name/config/samples/
```

## Key Differences: controller-runtime vs client-go

| Feature | client-go (main controller) | controller-runtime (examples) |
|---------|----------------------------|------------------------------|
| **Reconciliation** | Manual informer + workqueue | Automatic `Reconcile()` loop |
| **Resource Watching** | Manual `AddEventHandler` | Declarative `Watches()` |
| **Caching** | Manual informer cache | Built-in manager cache |
| **Status Updates** | Manual status subresource calls | `Status().Update()` |
| **Finalizers** | Manual string manipulation | `controllerutil` helpers |
| **Testing** | Mock clients | `envtest` with real API server |
| **CRD Development** | Manual YAML | kubebuilder markers + codegen |
| **Webhooks** | Manual server setup | Built-in webhook server |

See [docs/CONTROLLER_RUNTIME_GUIDE.md](../docs/CONTROLLER_RUNTIME_GUIDE.md) for detailed comparison.

## Migration

Interested in migrating the main controller to controller-runtime? See [docs/MIGRATION_GUIDE.md](../docs/MIGRATION_GUIDE.md) for a step-by-step guide with side-by-side comparisons.

## Additional Resources

- [Controller-Runtime Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [CRD Development Guide](../docs/CRD_DEVELOPMENT.md)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

## FAQ

### Q: Should I use controller-runtime or client-go?

**A:** Use controller-runtime for most new controllers. Use client-go when:
- You need fine-grained control over informers and caching
- You're building a low-level Kubernetes component
- You're extending an existing client-go codebase

See [docs/CONTROLLER_RUNTIME_GUIDE.md](../docs/CONTROLLER_RUNTIME_GUIDE.md) for detailed guidance.

### Q: Can I run multiple examples simultaneously?

**A:** Yes, but ensure:
1. Each example uses different CRD group/versions
2. Controllers listen on different ports (metrics, health)
3. You have sufficient cluster resources

### Q: How do I debug controller issues?

**A:** Use these techniques:
1. Enable debug logging: `go run main.go --zap-log-level=debug`
2. Check events: `kubectl describe <resource>`
3. View controller logs
4. Use metrics endpoint: `curl http://localhost:8080/metrics`

### Q: Can I use these examples in production?

**A:** These examples are educational. For production:
1. Add proper error handling and retries
2. Implement comprehensive observability
3. Add RBAC manifests
4. Consider leader election for HA
5. Add proper logging and metrics
6. Review security best practices

## Contributing

Found an issue or have a suggestion? Please:
1. Check existing examples for similar patterns
2. Ensure your suggestion aligns with controller-runtime best practices
3. Follow the existing sample structure
4. Include tests and documentation

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.
