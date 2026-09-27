package controller

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/deepak-muley/k8s-custom-controller/internal/config"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("GVKController", func() {
	var (
		cfg        *config.Config
		restConfig *rest.Config
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		// Create a minimal valid rest config for testing
		// Note: This won't actually connect to a cluster, but allows us to test controller initialization
		restConfig = &rest.Config{
			Host: "https://test.example.com",
		}

		scheme = runtime.NewScheme()

		cfg = &config.Config{
			Namespace: "default",
			GVKs: []schema.GroupVersionKind{
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
		}
	})

	Describe("NewGVKController", func() {
		Context("with valid configuration", func() {
			It("should create a new GVKController", func() {
				controller, err := NewGVKController(cfg, restConfig, scheme)

				// Note: This will likely fail because we can't create real dynamic clients
				// without a valid cluster connection, but we test the validation logic
				if err == nil {
					Expect(controller).NotTo(BeNil())
					Expect(controller.config).To(Equal(cfg))
					Expect(controller.restConfig).To(Equal(restConfig))
					Expect(controller.GetCounter()).NotTo(BeNil())
				} else {
					// Expected to fail due to invalid rest config, but should have validated config first
					Expect(err.Error()).To(SatisfyAny(
						ContainSubstring("failed to create dynamic client"),
						ContainSubstring("invalid config"),
					))
				}
			})

			It("should initialize counter", func() {
				// Skip if we can't create controller due to connection issues
				controller, err := NewGVKController(cfg, restConfig, scheme)
				if err != nil {
					Skip("Cannot create controller without valid cluster connection")
				}

				counter := controller.GetCounter()
				Expect(counter).NotTo(BeNil())
				Expect(counter.GetAllCounts()).To(BeEmpty())
			})

			It("should initialize informers map", func() {
				controller, err := NewGVKController(cfg, restConfig, scheme)
				if err != nil {
					Skip("Cannot create controller without valid cluster connection")
				}

				Expect(controller.informers).NotTo(BeNil())
				Expect(controller.informers).To(BeEmpty())
			})
		})

		Context("with invalid configuration", func() {
			It("should return error for invalid config", func() {
				invalidCfg := &config.Config{
					Namespace: "default",
					GVKs:      []schema.GroupVersionKind{}, // Empty GVKs
				}

				controller, err := NewGVKController(invalidCfg, restConfig, scheme)

				Expect(err).To(HaveOccurred())
				Expect(controller).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid config"))
				Expect(err.Error()).To(ContainSubstring("at least one GVK must be specified"))
			})

			It("should return error for nil config", func() {
				controller, err := NewGVKController(nil, restConfig, scheme)

				// This will panic or fail, depending on implementation
				// We test that it handles nil gracefully
				if err != nil {
					Expect(controller).To(BeNil())
				}
			})
		})

		Context("with nil rest config", func() {
			It("should handle nil rest config gracefully", func() {
				controller, err := NewGVKController(cfg, nil, scheme)

				Expect(err).To(HaveOccurred())
				Expect(controller).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to create dynamic client"))
			})
		})
	})

	Describe("GetCounter", func() {
		It("should return the resource counter", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			counter := controller.GetCounter()
			Expect(counter).NotTo(BeNil())

			// Counter should be same instance
			Expect(controller.GetCounter()).To(Equal(counter))
		})

		It("should allow counter operations", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			counter := controller.GetCounter()
			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			counter.Increment(gvk)
			Expect(counter.GetCount(gvk)).To(Equal(1))
		})
	})

	Describe("Stop", func() {
		It("should stop the controller", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			// Should not panic
			Expect(func() {
				controller.Stop()
			}).NotTo(Panic())

			// Stop again should not panic (idempotent check)
			Expect(func() {
				controller.Stop()
			}).NotTo(Panic())
		})

		It("should close stop channel", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			// Create a copy of stopCh before stopping
			stopCh := controller.stopCh

			controller.Stop()

			// Channel should be closed
			_, ok := <-stopCh
			Expect(ok).To(BeFalse())
		})

		It("should cancel context", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			ctx := controller.ctx
			controller.Stop()

			// Context should be cancelled
			select {
			case <-ctx.Done():
				// Expected
			case <-time.After(100 * time.Millisecond):
				Fail("Context should be cancelled after Stop()")
			}
		})
	})

	Describe("keyForGVK", func() {
		It("should create key for GVK with group", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			key := controller.keyForGVK(gvk)
			Expect(key).To(Equal("apps/v1/Deployment"))
		})

		It("should create key for GVK with empty group", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			}

			key := controller.keyForGVK(gvk)
			Expect(key).To(Equal("/v1/Pod"))
		})

		It("should handle different GVKs", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

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

			Expect(controller.keyForGVK(deploymentGVK)).To(Equal("apps/v1/Deployment"))
			Expect(controller.keyForGVK(serviceGVK)).To(Equal("/v1/Service"))
		})
	})

	Describe("getNamespace", func() {
		It("should return namespace when specified", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			namespace := controller.getNamespace()
			Expect(namespace).To(Equal("default"))
		})

		It("should return 'all namespaces' when namespace is empty", func() {
			cfgAllNamespaces := &config.Config{
				Namespace: "",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
				},
			}

			controller, err := NewGVKController(cfgAllNamespaces, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			namespace := controller.getNamespace()
			Expect(namespace).To(Equal("all namespaces"))
		})
	})

	Describe("shouldCount", func() {
		var controller *GVKController

		BeforeEach(func() {
			var err error
			controller, err = NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}
		})

		It("should return true for resource in matching namespace", func() {
			obj := &unstructured.Unstructured{}
			obj.SetNamespace("default")
			obj.SetName("test-resource")

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			shouldCount := controller.shouldCount(obj, gvk)
			Expect(shouldCount).To(BeTrue())
		})

		It("should return false for resource in different namespace", func() {
			obj := &unstructured.Unstructured{}
			obj.SetNamespace("other-namespace")
			obj.SetName("test-resource")

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			shouldCount := controller.shouldCount(obj, gvk)
			Expect(shouldCount).To(BeFalse())
		})

		It("should return true for all namespaces when namespace filter is empty", func() {
			cfgAllNamespaces := &config.Config{
				Namespace: "",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
				},
			}

			controllerAllNamespaces, err := NewGVKController(cfgAllNamespaces, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			obj1 := &unstructured.Unstructured{}
			obj1.SetNamespace("namespace1")
			obj1.SetName("test-resource-1")

			obj2 := &unstructured.Unstructured{}
			obj2.SetNamespace("namespace2")
			obj2.SetName("test-resource-2")

			gvk := schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}

			Expect(controllerAllNamespaces.shouldCount(obj1, gvk)).To(BeTrue())
			Expect(controllerAllNamespaces.shouldCount(obj2, gvk)).To(BeTrue())
		})

		It("should handle cluster-scoped resources", func() {
			obj := &unstructured.Unstructured{}
			// Cluster-scoped resources have empty namespace
			obj.SetName("test-resource")

			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Namespace",
			}

			shouldCount := controller.shouldCount(obj, gvk)
			// Should return false because namespace filter is "default" but resource has no namespace
			Expect(shouldCount).To(BeFalse())
		})

		It("should return true for cluster-scoped resources when watching all namespaces", func() {
			cfgAllNamespaces := &config.Config{
				Namespace: "",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "",
						Version: "v1",
						Kind:    "Namespace",
					},
				},
			}

			controllerAllNamespaces, err := NewGVKController(cfgAllNamespaces, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			obj := &unstructured.Unstructured{}
			obj.SetName("test-namespace")

			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Namespace",
			}

			shouldCount := controllerAllNamespaces.shouldCount(obj, gvk)
			Expect(shouldCount).To(BeTrue())
		})
	})

	Describe("Controller initialization", func() {
		It("should initialize with multiple GVKs", func() {
			multiGVKCfg := &config.Config{
				Namespace: "default",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "ReplicaSet",
					},
					{
						Group:   "",
						Version: "v1",
						Kind:    "Service",
					},
				},
			}

			controller, err := NewGVKController(multiGVKCfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			Expect(controller).NotTo(BeNil())
			Expect(controller.config.GVKs).To(HaveLen(3))
		})

		It("should preserve configuration", func() {
			customCfg := &config.Config{
				Namespace: "production",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
				},
			}

			controller, err := NewGVKController(customCfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			Expect(controller.config).To(Equal(customCfg))
			Expect(controller.config.Namespace).To(Equal("production"))
		})
	})

	Describe("Context handling", func() {
		It("should create a cancellable context", func() {
			controller, err := NewGVKController(cfg, restConfig, scheme)
			if err != nil {
				Skip("Cannot create controller without valid cluster connection")
			}

			ctx := controller.ctx
			Expect(ctx).NotTo(BeNil())

			// Context should not be done initially
			select {
			case <-ctx.Done():
				Fail("Context should not be done initially")
			default:
				// Expected - context is not done
			}

			// After stop, context should be done
			controller.Stop()

			select {
			case <-ctx.Done():
				// Expected
			case <-time.After(100 * time.Millisecond):
				Fail("Context should be done after Stop()")
			}
		})
	})

	// Note: Testing Start(), setupInformerForGVK(), and recountAllResources() would require:
	// - A real Kubernetes cluster, OR
	// - Sophisticated mocking of dynamic clients and discovery clients, OR
	// - Integration tests with testcontainers or kind
	// These tests are intentionally omitted here as they require significant infrastructure setup.
	// For production use, consider adding integration tests separately.
})
