package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/deepak-muley/k8s-custom-controller/internal/config"
	"github.com/deepak-muley/k8s-custom-controller/internal/counter"
)

// GVKController manages watching and counting resources for multiple GVKs
type GVKController struct {
	config          *config.Config
	restConfig      *rest.Config
	counter         *counter.ResourceCounter
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	informers       map[string]cache.SharedInformer
	informersLock   sync.RWMutex
	stopCh          chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewGVKController creates a new GVKController
func NewGVKController(cfg *config.Config, restConfig *rest.Config, scheme *runtime.Scheme) (*GVKController, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &GVKController{
		config:          cfg,
		restConfig:      restConfig,
		counter:         counter.NewResourceCounter(),
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		informers:       make(map[string]cache.SharedInformer),
		stopCh:          make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start starts the controller
func (gc *GVKController) Start(ctx context.Context) error {
	klog.Info("Starting GVK controller...")
	klog.Infof("Watching namespace: %s", gc.getNamespace())
	klog.Infof("Watching %d GVK(s)", len(gc.config.GVKs))

	// Create dynamic informer factory
	var informerFactory dynamicinformer.DynamicSharedInformerFactory
	if gc.config.Namespace != "" {
		informerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			gc.dynamicClient,
			time.Second*30,
			gc.config.Namespace,
			func(opts *metav1.ListOptions) {},
		)
	} else {
		informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(
			gc.dynamicClient,
			time.Second*30,
		)
	}

	// Setup informers for each GVK
	for _, gvk := range gc.config.GVKs {
		if err := gc.setupInformerForGVK(gvk, informerFactory); err != nil {
			return fmt.Errorf("failed to setup informer for GVK %s: %w", gvk, err)
		}
	}

	// Start all informers
	informerFactory.Start(gc.stopCh)

	// Wait for caches to sync
	klog.Info("Waiting for informer caches to sync...")
	if !cache.WaitForCacheSync(gc.stopCh, func() bool {
		gc.informersLock.RLock()
		defer gc.informersLock.RUnlock()
		for _, informer := range gc.informers {
			if !informer.HasSynced() {
				return false
			}
		}
		return true
	}) {
		return fmt.Errorf("failed to sync informer caches")
	}

	klog.Info("All informer caches synced successfully")

	// Initial count from existing resources
	if err := gc.recountAllResources(ctx); err != nil {
		klog.Warningf("Failed to perform initial count: %v", err)
	}

	// Start periodic counter display
	go gc.periodicDisplay(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	gc.Stop()
	return nil
}

// Stop stops the controller
func (gc *GVKController) Stop() {
	klog.Info("Stopping GVK controller...")
	gc.cancel()
	close(gc.stopCh)
}

// GetCounter returns the resource counter
func (gc *GVKController) GetCounter() *counter.ResourceCounter {
	return gc.counter
}

// setupInformerForGVK sets up an informer for a specific GVK
func (gc *GVKController) setupInformerForGVK(gvk schema.GroupVersionKind, factory dynamicinformer.DynamicSharedInformerFactory) error {
	// Create GVR from GVK - will be corrected via discovery
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "", // Will be set via discovery
	}

	// Get the correct resource name from discovery
	// For core resources (empty group), GroupVersion().String() returns "v1"
	groupVersion := gvk.GroupVersion().String()
	resources, err := gc.discoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return fmt.Errorf("failed to discover resources for %s: %w", groupVersion, err)
	}

	// Find the correct resource name by matching Kind
	found := false
	for _, resource := range resources.APIResources {
		if resource.Kind == gvk.Kind {
			gvr.Resource = resource.Name
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("resource kind %s not found in API group %s/%s", gvk.Kind, gvk.Group, gvk.Version)
	}

	// Create informer
	informer := factory.ForResource(gvr).Informer()

	// Set up event handlers
	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if u, ok := obj.(*unstructured.Unstructured); ok {
				if gc.shouldCount(u, gvk) {
					gc.counter.Increment(gvk)
					klog.Infof("Resource added: %s/%s (GVK: %s/%s/%s). Count: %d",
						u.GetNamespace(), u.GetName(), gvk.Group, gvk.Version, gvk.Kind, gc.counter.GetCount(gvk))
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Updates don't change the count, but we could log them if needed
			if u, ok := newObj.(*unstructured.Unstructured); ok {
				if gc.shouldCount(u, gvk) {
					// Count remains the same on update, but we could track changes
					klog.V(5).Infof("Resource updated: %s/%s (GVK: %s/%s/%s)",
						u.GetNamespace(), u.GetName(), gvk.Group, gvk.Version, gvk.Kind)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			var u *unstructured.Unstructured
			switch t := obj.(type) {
			case *unstructured.Unstructured:
				u = t
			case cache.DeletedFinalStateUnknown:
				if deletedObj, ok := t.Obj.(*unstructured.Unstructured); ok {
					u = deletedObj
				}
			}

			if u != nil && gc.shouldCount(u, gvk) {
				gc.counter.Decrement(gvk)
				klog.Infof("Resource deleted: %s/%s (GVK: %s/%s/%s). Count: %d",
					u.GetNamespace(), u.GetName(), gvk.Group, gvk.Version, gvk.Kind, gc.counter.GetCount(gvk))
			}
		},
	})

	if err != nil {
		return fmt.Errorf("failed to add event handlers: %w", err)
	}

	// Store the informer
	key := gc.keyForGVK(gvk)
	gc.informersLock.Lock()
	gc.informers[key] = informer
	gc.informersLock.Unlock()

	klog.Infof("Registered informer for GVK: %s/%s/%s (Resource: %s)",
		gvk.Group, gvk.Version, gvk.Kind, gvr.Resource)
	return nil
}

// shouldCount determines if a resource should be counted based on namespace filter
func (gc *GVKController) shouldCount(obj *unstructured.Unstructured, gvk schema.GroupVersionKind) bool {
	// Check namespace filter
	if gc.config.Namespace != "" {
		if obj.GetNamespace() != gc.config.Namespace {
			return false
		}
	}
	return true
}

// recountAllResources performs an initial count of all existing resources
func (gc *GVKController) recountAllResources(ctx context.Context) error {
	klog.Info("Performing initial count of existing resources...")

	// Reset counter
	gc.counter.Reset()

	for _, gvk := range gc.config.GVKs {
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: "",
		}

		// Get correct resource name from discovery
		groupVersion := gvk.GroupVersion().String()
		resources, discoverErr := gc.discoveryClient.ServerResourcesForGroupVersion(groupVersion)
		if discoverErr != nil {
			klog.Warningf("Failed to discover resources for %s: %v", groupVersion, discoverErr)
			continue
		}

		for _, resource := range resources.APIResources {
			if resource.Kind == gvk.Kind {
				gvr.Resource = resource.Name
				break
			}
		}

		if gvr.Resource == "" {
			klog.Warningf("Resource kind %s not found for %s/%s", gvk.Kind, gvk.Group, gvk.Version)
			continue
		}

		var list *unstructured.UnstructuredList
		var err error

		if gc.config.Namespace != "" {
			list, err = gc.dynamicClient.Resource(gvr).Namespace(gc.config.Namespace).List(ctx, metav1.ListOptions{})
		} else {
			list, err = gc.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			klog.Warningf("Failed to list resources for GVK %s/%s/%s: %v", gvk.Group, gvk.Version, gvk.Kind, err)
			continue
		}

		count := len(list.Items)
		// Set count directly for initial count
		gc.counter.SetCount(gvk, count)

		klog.Infof("Initial count for %s/%s/%s: %d", gvk.Group, gvk.Version, gvk.Kind, count)
	}

	klog.Info("Initial count completed")
	return nil
}

// periodicDisplay displays counter information periodically
func (gc *GVKController) periodicDisplay(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			klog.Info(gc.counter.String())
		}
	}
}

// keyForGVK creates a string key for a GVK
func (gc *GVKController) keyForGVK(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
}

// getNamespace returns the namespace string (or "all namespaces" if empty)
func (gc *GVKController) getNamespace() string {
	if gc.config.Namespace == "" {
		return "all namespaces"
	}
	return gc.config.Namespace
}
