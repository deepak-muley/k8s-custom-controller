package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"github.com/deepak-muley/k8s-custom-controller/internal/config"
)

// This file contains integration tests using envtest.
// envtest provides a real Kubernetes API server (etcd + API server) for testing.
// Reference: https://book.kubebuilder.io/cronjob-tutorial/writing-tests

var _ = Describe("GVKController with envtest", func() {
	var (
		controller       *GVKController
		dynamicK8sClient *kubernetes.Clientset
		testCtx          context.Context
		testCancel       context.CancelFunc
		deploymentGVK    schema.GroupVersionKind
		serviceGVK       schema.GroupVersionKind
	)

	BeforeEach(func() {
		// Create a context for this test
		testCtx, testCancel = context.WithCancel(context.Background())

		// Create a Kubernetes clientset for creating test resources
		var err error
		dynamicK8sClient, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		// Create test namespace
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: TestNamespace,
			},
		}
		_, err = dynamicK8sClient.CoreV1().Namespaces().Create(testCtx, namespace, metav1.CreateOptions{})
		// Ignore error if namespace already exists
		if err != nil {
			GinkgoWriter.Printf("Note: namespace may already exist: %v\n", err)
		}

		// Define GVKs we'll test with
		deploymentGVK = schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}
		serviceGVK = schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Service",
		}
	})

	AfterEach(func() {
		if controller != nil {
			controller.Stop()
		}
		if testCancel != nil {
			testCancel()
		}

		// Cleanup: Delete test namespace
		if dynamicK8sClient != nil {
			err := dynamicK8sClient.CoreV1().Namespaces().Delete(testCtx, TestNamespace, metav1.DeleteOptions{})
			if err != nil {
				GinkgoWriter.Printf("Note: error cleaning up namespace: %v\n", err)
			}
		}
	})

	Context("Controller Discovery and Setup", func() {
		It("should create controller with valid config", func() {
			By("creating a controller config")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())
			Expect(controller).NotTo(BeNil())
			Expect(controller.GetCounter()).NotTo(BeNil())
		})

		It("should discover resources correctly using discovery API", func() {
			By("creating a controller")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			controller, err := NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("querying discovery API for apps/v1 resources")
			// This tests that the discovery client works with envtest API server
			resources, err := controller.discoveryClient.ServerResourcesForGroupVersion("apps/v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).NotTo(BeNil())

			By("verifying deployments resource is discovered")
			found := false
			for _, resource := range resources.APIResources {
				if resource.Kind == "Deployment" && resource.Name == "deployments" {
					found = true
					GinkgoWriter.Printf("Found deployments resource: %+v\n", resource)
					break
				}
			}
			Expect(found).To(BeTrue(), "Should discover deployments resource via discovery API")
		})

		It("should discover core v1 resources", func() {
			By("creating a controller for core resources")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{serviceGVK},
			}

			scheme := runtime.NewScheme()
			controller, err := NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("querying discovery API for v1 resources")
			resources, err := controller.discoveryClient.ServerResourcesForGroupVersion("v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).NotTo(BeNil())

			By("verifying services resource is discovered")
			found := false
			for _, resource := range resources.APIResources {
				if resource.Kind == "Service" && resource.Name == "services" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Should discover services resource via discovery API")
		})
	})

	Context("Controller Lifecycle", func() {
		It("should start controller and sync informer cache", func() {
			By("creating a controller")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("starting the controller")
			started := make(chan error, 1)
			go func() {
				started <- controller.Start(testCtx)
			}()

			By("waiting for informer cache to sync")
			Eventually(func(g Gomega) {
				controller.informersLock.RLock()
				defer controller.informersLock.RUnlock()
				for _, informer := range controller.informers {
					g.Expect(informer.HasSynced()).To(BeTrue(), "All informers should be synced")
				}
				g.Expect(len(controller.informers)).To(BeNumerically(">", 0), "Should have at least one informer")
			}, timeout, interval).Should(Succeed(), "Informer cache should sync within timeout")

			// Verify controller is still running (hasn't errored)
			select {
			case err := <-started:
				if err != nil {
					Fail("Controller should not have errored: " + err.Error())
				}
			case <-time.After(100 * time.Millisecond):
				// Controller is still running, which is expected
			}
		})

		It("should stop controller gracefully", func() {
			By("creating and starting a controller")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			// Start controller
			go func() {
				_ = controller.Start(testCtx)
			}()

			// Wait a bit for it to start
			time.Sleep(500 * time.Millisecond)

			By("stopping the controller")
			controller.Stop()

			By("verifying context is cancelled")
			Eventually(func() bool {
				select {
				case <-controller.ctx.Done():
					return true
				default:
					return false
				}
			}, 5*time.Second, 100*time.Millisecond).Should(BeTrue(), "Controller context should be cancelled after Stop()")
		})
	})

	Context("Resource Counting", func() {
		It("should count resources when they are created", func() {
			By("creating a controller")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("starting the controller")
			go func() {
				_ = controller.Start(testCtx)
			}()

			By("waiting for informer cache to sync")
			Eventually(func() bool {
				controller.informersLock.RLock()
				defer controller.informersLock.RUnlock()
				for _, informer := range controller.informers {
					if !informer.HasSynced() {
						return false
					}
				}
				return len(controller.informers) > 0
			}, timeout, interval).Should(BeTrue())

			By("creating a test deployment using dynamic client")
			deployment := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      DeploymentName,
						"namespace": TestNamespace,
					},
					"spec": map[string]interface{}{
						"replicas": int64(1),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"app": "test",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "test",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			}

			// Create deployment using dynamic client
			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}
			createdDeployment, err := controller.dynamicClient.Resource(gvr).Namespace(TestNamespace).Create(testCtx, deployment, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdDeployment).NotTo(BeNil())

			By("waiting for controller to process the create event and increment count")
			Eventually(func(g Gomega) {
				count := controller.counter.GetCount(deploymentGVK)
				g.Expect(count).To(BeNumerically(">=", 1), "Should count at least one deployment")
			}, timeout, interval).Should(Succeed(), "Controller should count created deployment")

			By("verifying the count is correct")
			finalCount := controller.counter.GetCount(deploymentGVK)
			Expect(finalCount).To(BeNumerically(">=", 1))
			GinkgoWriter.Printf("Final deployment count: %d\n", finalCount)
		})

		It("should handle resource deletion and decrement count", func() {
			By("creating a controller")
			controllerCfg := &config.Config{
				Namespace: TestNamespace,
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("starting the controller")
			go func() {
				_ = controller.Start(testCtx)
			}()

			By("waiting for informer cache to sync")
			Eventually(func() bool {
				controller.informersLock.RLock()
				defer controller.informersLock.RUnlock()
				for _, informer := range controller.informers {
					if !informer.HasSynced() {
						return false
					}
				}
				return len(controller.informers) > 0
			}, timeout, interval).Should(BeTrue())

			By("creating a test deployment")
			deployment := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      DeploymentName + "-delete",
						"namespace": TestNamespace,
					},
					"spec": map[string]interface{}{
						"replicas": int64(1),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"app": "test",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "test",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			}

			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}
			_, err = controller.dynamicClient.Resource(gvr).Namespace(TestNamespace).Create(testCtx, deployment, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("waiting for deployment to be counted")
			Eventually(func() int {
				return controller.counter.GetCount(deploymentGVK)
			}, timeout, interval).Should(BeNumerically(">=", 1))

			initialCount := controller.counter.GetCount(deploymentGVK)

			By("deleting the deployment")
			err = controller.dynamicClient.Resource(gvr).Namespace(TestNamespace).Delete(testCtx, DeploymentName+"-delete", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("waiting for controller to process the delete event and decrement count")
			Eventually(func(g Gomega) {
				count := controller.counter.GetCount(deploymentGVK)
				g.Expect(count).To(BeNumerically("<", initialCount), "Count should decrease after deletion")
			}, timeout, interval).Should(Succeed(), "Controller should decrement count after deletion")

			By("verifying the count decreased")
			finalCount := controller.counter.GetCount(deploymentGVK)
			Expect(finalCount).To(BeNumerically("<", initialCount))
			GinkgoWriter.Printf("Initial count: %d, Final count after deletion: %d\n", initialCount, finalCount)
		})

		It("should filter resources by namespace", func() {
			By("creating a controller with namespace filter")
			controllerCfg := &config.Config{
				Namespace: TestNamespace, // Watch only this namespace
				GVKs:      []schema.GroupVersionKind{deploymentGVK},
			}

			scheme := runtime.NewScheme()
			var err error
			controller, err = NewGVKController(controllerCfg, cfg, scheme)
			Expect(err).NotTo(HaveOccurred())

			By("creating another namespace")
			otherNamespace := TestNamespace + "-other"
			otherNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: otherNamespace,
				},
			}
			_, err = dynamicK8sClient.CoreV1().Namespaces().Create(testCtx, otherNs, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			defer dynamicK8sClient.CoreV1().Namespaces().Delete(testCtx, otherNamespace, metav1.DeleteOptions{})

			By("starting the controller")
			go func() {
				_ = controller.Start(testCtx)
			}()

			By("waiting for informer cache to sync")
			Eventually(func() bool {
				controller.informersLock.RLock()
				defer controller.informersLock.RUnlock()
				for _, informer := range controller.informers {
					if !informer.HasSynced() {
						return false
					}
				}
				return len(controller.informers) > 0
			}, timeout, interval).Should(BeTrue())

			By("creating a deployment in the watched namespace")
			deploymentInWatched := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      DeploymentName + "-watched",
						"namespace": TestNamespace,
					},
					"spec": map[string]interface{}{
						"replicas": int64(1),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"app": "test",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "test",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			}

			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}
			_, err = controller.dynamicClient.Resource(gvr).Namespace(TestNamespace).Create(testCtx, deploymentInWatched, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("creating a deployment in another namespace (should not be counted)")
			deploymentInOther := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      DeploymentName + "-other",
						"namespace": otherNamespace,
					},
					"spec": map[string]interface{}{
						"replicas": int64(1),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"app": "test",
							},
						},
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"app": "test",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "test",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
			}

			_, createErr := controller.dynamicClient.Resource(gvr).Namespace(otherNamespace).Create(testCtx, deploymentInOther, metav1.CreateOptions{})
			Expect(createErr).NotTo(HaveOccurred())

			By("verifying only the watched namespace deployment is counted")
			Eventually(func(g Gomega) {
				count := controller.counter.GetCount(deploymentGVK)
				g.Expect(count).To(Equal(1), "Should only count deployment in watched namespace")
			}, timeout, interval).Should(Succeed(), "Namespace filtering should work correctly")
		})
	})
})
