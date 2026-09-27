package counter

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCounter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Counter Suite")
}

var _ = Describe("ResourceCounter", func() {
	var counter *ResourceCounter

	BeforeEach(func() {
		counter = NewResourceCounter()
	})

	Describe("NewResourceCounter", func() {
		It("should create a new ResourceCounter with empty counts", func() {
			Expect(counter).NotTo(BeNil())
			Expect(counter.GetAllCounts()).To(BeEmpty())
			Expect(counter.GetLastUpdated()).To(BeZero())
		})
	})

	Describe("Increment", func() {
		It("should increment count for a given GVK from zero to one", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)

			Expect(counter.GetCount(gvk)).To(Equal(1))
		})

		It("should increment count multiple times", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			counter.Increment(gvk)
			counter.Increment(gvk)

			Expect(counter.GetCount(gvk)).To(Equal(3))
		})

		It("should update lastUpdated timestamp", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			beforeTime := time.Now()
			time.Sleep(10 * time.Millisecond) // Ensure time difference
			counter.Increment(gvk)
			afterTime := time.Now()

			lastUpdated := counter.GetLastUpdated()
			Expect(lastUpdated).To(BeTemporally(">=", beforeTime))
			Expect(lastUpdated).To(BeTemporally("<=", afterTime))
		})

		It("should handle core API group (empty group)", func() {
			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			}

			counter.Increment(gvk)

			Expect(counter.GetCount(gvk)).To(Equal(1))
			counts := counter.GetAllCounts()
			Expect(counts).To(HaveKey("core/v1/Pod"))
		})

		It("should handle multiple different GVKs independently", func() {
			deploymentGVK := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			serviceGVK := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			}

			counter.Increment(deploymentGVK)
			counter.Increment(deploymentGVK)
			counter.Increment(serviceGVK)

			Expect(counter.GetCount(deploymentGVK)).To(Equal(2))
			Expect(counter.GetCount(serviceGVK)).To(Equal(1))
		})
	})

	Describe("Decrement", func() {
		It("should not decrement below zero", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Decrement(gvk)

			Expect(counter.GetCount(gvk)).To(Equal(0))
		})

		It("should decrement count correctly", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			counter.Increment(gvk)
			counter.Increment(gvk)
			counter.Decrement(gvk)

			Expect(counter.GetCount(gvk)).To(Equal(2))
		})

		It("should update lastUpdated timestamp on decrement", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			firstUpdate := counter.GetLastUpdated()

			time.Sleep(10 * time.Millisecond)
			counter.Decrement(gvk)

			Expect(counter.GetLastUpdated()).To(BeTemporally(">", firstUpdate))
		})
	})

	Describe("SetCount", func() {
		It("should set count to a specific value", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.SetCount(gvk, 5)

			Expect(counter.GetCount(gvk)).To(Equal(5))
		})

		It("should overwrite existing count", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			counter.Increment(gvk)
			counter.SetCount(gvk, 10)

			Expect(counter.GetCount(gvk)).To(Equal(10))
		})

		It("should handle zero count", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.SetCount(gvk, 5)
			counter.SetCount(gvk, 0)

			Expect(counter.GetCount(gvk)).To(Equal(0))
		})

		It("should update lastUpdated timestamp", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			firstUpdate := counter.GetLastUpdated()

			time.Sleep(10 * time.Millisecond)
			counter.SetCount(gvk, 10)

			Expect(counter.GetLastUpdated()).To(BeTemporally(">", firstUpdate))
		})
	})

	Describe("GetCount", func() {
		It("should return zero for non-existent GVK", func() {
			gvk := schema.GroupVersionKind{
				Group:   "nonexistent",
				Version: "v1",
				Kind:    "Resource",
			}

			Expect(counter.GetCount(gvk)).To(Equal(0))
		})

		It("should return current count for existing GVK", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.SetCount(gvk, 7)

			Expect(counter.GetCount(gvk)).To(Equal(7))
		})
	})

	Describe("GetAllCounts", func() {
		It("should return empty map for new counter", func() {
			counts := counter.GetAllCounts()
			Expect(counts).To(BeEmpty())
		})

		It("should return copy of all counts", func() {
			deploymentGVK := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			serviceGVK := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			}

			counter.SetCount(deploymentGVK, 5)
			counter.SetCount(serviceGVK, 3)

			counts := counter.GetAllCounts()
			Expect(counts).To(HaveLen(2))
			Expect(counts).To(HaveKey("apps/v1/Deployment"))
			Expect(counts).To(HaveKey("core/v1/Service"))
			Expect(counts["apps/v1/Deployment"]).To(Equal(5))
			Expect(counts["core/v1/Service"]).To(Equal(3))
		})

		It("should return a copy that doesn't affect original", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.SetCount(gvk, 5)
			counts := counter.GetAllCounts()
			counts["apps/v1/Deployment"] = 99 // Modify the copy

			Expect(counter.GetCount(gvk)).To(Equal(5)) // Original unchanged
		})
	})

	Describe("GetLastUpdated", func() {
		It("should return zero time for new counter", func() {
			Expect(counter.GetLastUpdated()).To(BeZero())
		})

		It("should return last update time after operations", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			beforeTime := time.Now()
			counter.Increment(gvk)
			lastUpdated := counter.GetLastUpdated()
			afterTime := time.Now()

			Expect(lastUpdated).To(BeTemporally(">=", beforeTime))
			Expect(lastUpdated).To(BeTemporally("<=", afterTime))
		})
	})

	Describe("String", func() {
		It("should return message for empty counter", func() {
			result := counter.String()
			Expect(result).To(Equal("No resources tracked yet"))
		})

		It("should format counts correctly", func() {
			deploymentGVK := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			serviceGVK := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			}

			counter.SetCount(deploymentGVK, 5)
			counter.SetCount(serviceGVK, 3)

			result := counter.String()
			Expect(result).To(ContainSubstring("Resource Counts:"))
			Expect(result).To(ContainSubstring("Last Updated:"))
			Expect(result).To(ContainSubstring("apps/v1/Deployment"))
			Expect(result).To(ContainSubstring("core/v1/Service"))
			Expect(result).To(ContainSubstring("5"))
			Expect(result).To(ContainSubstring("3"))
		})

		It("should include RFC3339 formatted timestamp", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.SetCount(gvk, 1)
			result := counter.String()

			// Check that timestamp is in RFC3339 format
			Expect(result).To(MatchRegexp(`Last Updated: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`))
		})
	})

	Describe("Reset", func() {
		It("should clear all counts", func() {
			deploymentGVK := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			serviceGVK := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			}

			counter.SetCount(deploymentGVK, 5)
			counter.SetCount(serviceGVK, 3)

			counter.Reset()

			Expect(counter.GetAllCounts()).To(BeEmpty())
			Expect(counter.GetCount(deploymentGVK)).To(Equal(0))
			Expect(counter.GetCount(serviceGVK)).To(Equal(0))
		})

		It("should update lastUpdated timestamp", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			firstUpdate := counter.GetLastUpdated()

			time.Sleep(10 * time.Millisecond)
			counter.Reset()

			Expect(counter.GetLastUpdated()).To(BeTemporally(">", firstUpdate))
		})
	})

	Describe("Thread Safety", func() {
		It("should handle concurrent increments safely", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			done := make(chan bool)
			goroutines := 10
			incrementsPerGoroutine := 10

			for i := 0; i < goroutines; i++ {
				go func() {
					for j := 0; j < incrementsPerGoroutine; j++ {
						counter.Increment(gvk)
					}
					done <- true
				}()
			}

			for i := 0; i < goroutines; i++ {
				<-done
			}

			expectedCount := goroutines * incrementsPerGoroutine
			Expect(counter.GetCount(gvk)).To(Equal(expectedCount))
		})

		It("should handle concurrent read and write operations safely", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			done := make(chan bool)

			// Writer goroutines
			for i := 0; i < 5; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						counter.Increment(gvk)
						counter.Decrement(gvk)
					}
					done <- true
				}()
			}

			// Reader goroutines
			for i := 0; i < 5; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						_ = counter.GetCount(gvk)
						_ = counter.GetAllCounts()
						_ = counter.String()
					}
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			// Should not have panicked and should have valid state
			count := counter.GetCount(gvk)
			Expect(count).To(BeNumerically(">=", 0))
		})
	})

	Describe("keyForGVK", func() {
		It("should format core API group with 'core' prefix", func() {
			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			}

			counter.Increment(gvk)
			counts := counter.GetAllCounts()

			Expect(counts).To(HaveKey("core/v1/Pod"))
			Expect(counts).NotTo(HaveKey("/v1/Pod"))
		})

		It("should format named group correctly", func() {
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			counts := counter.GetAllCounts()

			Expect(counts).To(HaveKey("apps/v1/Deployment"))
		})
	})
})
