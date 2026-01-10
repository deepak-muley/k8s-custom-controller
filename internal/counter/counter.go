package counter

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceCounter maintains counts of resources by GVK and name
type ResourceCounter struct {
	mu          sync.RWMutex
	counts      map[string]int // key: "group/version/kind", value: count
	lastUpdated time.Time
}

// NewResourceCounter creates a new ResourceCounter
func NewResourceCounter() *ResourceCounter {
	return &ResourceCounter{
		counts: make(map[string]int),
	}
}

// Increment increments the count for a given GVK
func (rc *ResourceCounter) Increment(gvk schema.GroupVersionKind) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	key := rc.keyForGVK(gvk)
	rc.counts[key]++
	rc.lastUpdated = time.Now()
}

// Decrement decrements the count for a given GVK
func (rc *ResourceCounter) Decrement(gvk schema.GroupVersionKind) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	key := rc.keyForGVK(gvk)
	if rc.counts[key] > 0 {
		rc.counts[key]--
	}
	rc.lastUpdated = time.Now()
}

// SetCount sets the count directly for a given GVK
func (rc *ResourceCounter) SetCount(gvk schema.GroupVersionKind, count int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	key := rc.keyForGVK(gvk)
	rc.counts[key] = count
	rc.lastUpdated = time.Now()
}

// GetCount returns the current count for a given GVK
func (rc *ResourceCounter) GetCount(gvk schema.GroupVersionKind) int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	key := rc.keyForGVK(gvk)
	return rc.counts[key]
}

// GetAllCounts returns a copy of all counts
func (rc *ResourceCounter) GetAllCounts() map[string]int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	result := make(map[string]int, len(rc.counts))
	for k, v := range rc.counts {
		result[k] = v
	}
	return result
}

// GetLastUpdated returns the last update time
func (rc *ResourceCounter) GetLastUpdated() time.Time {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.lastUpdated
}

// String returns a formatted string representation of all counts
func (rc *ResourceCounter) String() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if len(rc.counts) == 0 {
		return "No resources tracked yet"
	}

	result := "Resource Counts:\n"
	result += fmt.Sprintf("Last Updated: %s\n\n", rc.lastUpdated.Format(time.RFC3339))

	for key, count := range rc.counts {
		result += fmt.Sprintf("  %s: %d\n", key, count)
	}

	return result
}

// keyForGVK creates a string key for a GVK
func (rc *ResourceCounter) keyForGVK(gvk schema.GroupVersionKind) string {
	if gvk.Group == "" {
		return fmt.Sprintf("core/%s/%s", gvk.Version, gvk.Kind)
	}
	return fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
}

// Reset resets all counts (useful for testing or reinitialization)
func (rc *ResourceCounter) Reset() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.counts = make(map[string]int)
	rc.lastUpdated = time.Now()
}
