# Example 09: Advanced Indexing

This example demonstrates how to use field indexing in controller-runtime for fast cache lookups. Field indexing allows you to efficiently query resources by custom fields without having to list all resources and filter them in memory.

## Key Concepts

### Field Indexing
Field indexing creates an in-memory index on the controller's cache for specific fields. This enables:
- **Fast lookups** - O(1) lookups instead of O(n) filtering
- **Efficient queries** - Query by custom spec fields
- **Multiple indexes** - Index multiple fields simultaneously
- **Automatic updates** - Index automatically updates as resources change

### Use Cases
- Finding resources by category or type
- Looking up resources by owner
- Querying by priority or importance
- Any custom field-based lookup pattern

## Architecture

```
┌─────────────────────────────────────────────┐
│          Manager with Field Indexes         │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │  Cache with Multiple Field Indexes  │   │
│  │                                     │   │
│  │  Index: spec.category               │   │
│  │    "database" -> [mysql, postgres]  │   │
│  │    "cache" -> [redis, memcached]    │   │
│  │                                     │   │
│  │  Index: spec.owner                  │   │
│  │    "team-backend" -> [mysql, ...]   │   │
│  │    "team-sre" -> [prometheus, ...]  │   │
│  │                                     │   │
│  │  Index: spec.priority               │   │
│  │    "10" -> [postgres, nginx, ...]   │   │
│  │    "9" -> [mysql, kafka, ...]       │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │    IndexedResource Controller       │   │
│  │                                     │   │
│  │  • Uses MatchingFields for lookups  │   │
│  │  • Fast category-based queries      │   │
│  │  • Owner-based resource finding     │   │
│  │  • Priority filtering               │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Features Demonstrated

### 1. Field Index Setup
```go
// In SetupWithManager
mgr.GetFieldIndexer().IndexField(
    context.Background(),
    &indexingv1alpha1.IndexedResource{},
    "spec.category",
    func(rawObj client.Object) []string {
        resource := rawObj.(*indexingv1alpha1.IndexedResource)
        return []string{resource.Spec.Category}
    },
)
```

### 2. Fast Cache Lookups
```go
// Using MatchingFields for indexed lookups
var resourceList indexingv1alpha1.IndexedResourceList
err := r.List(ctx, &resourceList,
    client.InNamespace(namespace),
    client.MatchingFields{"spec.category": "database"},
)
```

### 3. Multiple Field Indexes
The controller demonstrates indexing on:
- `spec.category` - String field for grouping resources
- `spec.owner` - String field for ownership tracking
- `spec.priority` - Integer field (indexed as string)

### 4. Integer Field Indexing
```go
// Integer fields are converted to strings for indexing
mgr.GetFieldIndexer().IndexField(
    context.Background(),
    &indexingv1alpha1.IndexedResource{},
    "spec.priority",
    func(rawObj client.Object) []string {
        resource := rawObj.(*indexingv1alpha1.IndexedResource)
        return []string{fmt.Sprintf("%d", resource.Spec.Priority)}
    },
)
```

## CRD Structure

### IndexedResource
```yaml
spec:
  category: database        # Indexed field for grouping
  owner: team-backend       # Indexed field for ownership
  priority: 9               # Indexed field for importance (1-10)
  tags:                     # Additional metadata
    - mysql
    - relational
  data: "Configuration..."  # Arbitrary payload

status:
  phase: Active
  lastProcessed: "2025-01-15T10:30:00Z"
  relatedResources:         # Found via index lookups
    - database-postgres
    - database-mongodb
  message: "Found 2 related resources in category 'database'"
```

## Running the Example

### 1. Install the CRD
```bash
kubectl apply -f config/crd/indexing.examples.k8s.io_indexedresources.yaml
```

### 2. Run the Controller
```bash
# From the repository root
./bin/09-advanced-indexing
```

### 3. Apply Sample Resources
```bash
kubectl apply -f config/samples/indexed_sample.yaml
```

### 4. Observe Index Usage
The controller will:
1. Find all resources in the same category
2. Look up resources by owner
3. Query resources by priority
4. Update status with related resources found

## Testing Index Queries

### Query by Category
```bash
# Create resources in the same category
kubectl apply -f config/samples/indexed_sample.yaml

# Check status to see related resources
kubectl get indexedresources database-mysql -o yaml
# Should show relatedResources: [database-postgres]
```

### Query by Owner
```bash
# All resources owned by team-backend
kubectl get indexedresources -o json | \
  jq '.items[] | select(.spec.owner=="team-backend") | .metadata.name'
```

### Query by Priority
```bash
# All high-priority resources (priority >= 9)
kubectl get indexedresources -o json | \
  jq '.items[] | select(.spec.priority >= 9) | .metadata.name'
```

## Performance Benefits

### Without Indexing
```go
// Must list ALL resources and filter in memory - O(n)
var allResources indexingv1alpha1.IndexedResourceList
r.List(ctx, &allResources, client.InNamespace(namespace))

var filtered []indexingv1alpha1.IndexedResource
for _, resource := range allResources.Items {
    if resource.Spec.Category == targetCategory {
        filtered = append(filtered, resource)
    }
}
```

### With Indexing
```go
// Direct cache lookup using index - O(1)
var resourceList indexingv1alpha1.IndexedResourceList
r.List(ctx, &resourceList,
    client.InNamespace(namespace),
    client.MatchingFields{"spec.category": targetCategory},
)
// resourceList.Items already contains only matching resources
```

## Best Practices

### 1. Index Setup
- Set up indexes in `SetupWithManager` before starting the controller
- Index fields that are frequently queried
- Keep index functions simple and deterministic

### 2. Field Selection
- Index fields with low cardinality (limited unique values)
- Avoid indexing frequently changing fields
- Consider query patterns in your controller logic

### 3. Memory Considerations
- Indexes consume memory proportional to the number of resources
- Each index maintains a map of field values to resources
- Balance query performance with memory usage

### 4. Index Functions
- Must return string slice (even for non-string fields)
- Should be deterministic (same input -> same output)
- Can return multiple values for multi-value fields

## Common Patterns

### Pattern 1: Category-Based Grouping
```go
// Find all resources in the same category
func (r *Reconciler) findRelatedResources(ctx context.Context, category string) ([]Resource, error) {
    var list ResourceList
    err := r.List(ctx, &list,
        client.MatchingFields{"spec.category": category},
    )
    return list.Items, err
}
```

### Pattern 2: Owner-Based Lookup
```go
// Find all resources owned by a team
func (r *Reconciler) findResourcesByOwner(ctx context.Context, owner string) ([]Resource, error) {
    var list ResourceList
    err := r.List(ctx, &list,
        client.MatchingFields{"spec.owner": owner},
    )
    return list.Items, err
}
```

### Pattern 3: Multi-Value Index
```go
// Index function returning multiple values
mgr.GetFieldIndexer().IndexField(
    context.Background(),
    &Resource{},
    "spec.tags",
    func(rawObj client.Object) []string {
        resource := rawObj.(*Resource)
        return resource.Spec.Tags  // Returns all tags
    },
)
```

## Cleanup

```bash
# Delete all sample resources
kubectl delete -f config/samples/indexed_sample.yaml

# Delete the CRD (this also deletes all IndexedResource instances)
kubectl delete -f config/crd/indexing.examples.k8s.io_indexedresources.yaml
```

## Key Takeaways

1. **Field indexing enables fast, efficient queries** on custom resource fields
2. **Set up indexes in SetupWithManager** before the controller starts
3. **Use MatchingFields** with indexed fields for O(1) lookups
4. **Index fields with low cardinality** that are frequently queried
5. **Integer and other types** must be converted to strings for indexing
6. **Indexes update automatically** as resources change in the cache

## Next Steps

- **Example 10**: Event Source Chaining - Watch multiple resource types
- **Example 11**: Rate Limiting & Backoff - Control reconciliation rate
- Review the [Controller Runtime Guide](../../docs/CONTROLLER_RUNTIME_GUIDE.md) for more patterns

## References

- [Controller Runtime Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client)
- [Field Indexer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#FieldIndexer)
- [Cache Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache)
