package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/deepak-muley/k8s-custom-controller/internal/config"
	"github.com/deepak-muley/k8s-custom-controller/internal/controller"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "Path to kubeconfig file (optional, defaults to in-cluster config)")
	namespace  = flag.String("namespace", "default", "Namespace to watch (empty string for all namespaces)")
	gvkFlag    = flag.String("gvk", "", "Comma-separated list of GVKs to watch in format 'group/version/kind' (e.g., 'apps/v1/Deployment,apps/v1/ReplicaSet')")
	logLevel   = flag.Int("v", 2, "Log level (higher = more verbose)")
)

func main() {
	flag.Parse()
	klog.InitFlags(nil)
	flag.Set("v", fmt.Sprintf("%d", *logLevel))

	// Build Kubernetes config
	restConfig, err := buildKubeConfig(*kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Parse GVKs
	gvks, err := parseGVKs(*gvkFlag)
	if err != nil {
		klog.Fatalf("Failed to parse GVKs: %v", err)
	}

	if len(gvks) == 0 {
		klog.Fatal("At least one GVK must be specified using -gvk flag")
	}

	// Create config
	cfg := &config.Config{
		Namespace: *namespace,
		GVKs:      gvks,
	}

	// Create controller
	k8sScheme := runtime.NewScheme()
	scheme.AddToScheme(k8sScheme)

	gvkController, err := controller.NewGVKController(cfg, restConfig, k8sScheme)
	if err != nil {
		klog.Fatalf("Failed to create controller: %v", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start controller in a goroutine
	errChan := make(chan error, 1)
	go func() {
		klog.Info("Starting controller...")
		if err := gvkController.Start(ctx); err != nil {
			errChan <- fmt.Errorf("controller error: %w", err)
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		klog.Infof("Received signal %v, shutting down...", sig)
		cancel()
		// Give controller time to shut down gracefully
		time.Sleep(2 * time.Second)
	case err := <-errChan:
		klog.Fatalf("Controller failed: %v", err)
	}

	// Print final counts
	klog.Info("\n=== Final Resource Counts ===")
	klog.Info(gvkController.GetCounter().String())
	klog.Info("Controller stopped")
}

// buildKubeConfig builds a Kubernetes REST config from kubeconfig path or in-cluster config
func buildKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig file: %w", err)
		}
		return config, nil
	}

	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fallback to default kubeconfig location
	config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	return config, nil
}

// parseGVKs parses comma-separated GVK strings into schema.GroupVersionKind objects
// Format: "group/version/kind,group2/version2/kind2"
func parseGVKs(gvkString string) ([]schema.GroupVersionKind, error) {
	if gvkString == "" {
		return nil, fmt.Errorf("GVK string cannot be empty")
	}

	var gvks []schema.GroupVersionKind
	parts := splitAndTrim(gvkString, ",")

	for _, part := range parts {
		gvk, err := parseGVK(part)
		if err != nil {
			return nil, fmt.Errorf("invalid GVK format '%s': %w", part, err)
		}
		gvks = append(gvks, gvk)
	}

	return gvks, nil
}

// parseGVK parses a single GVK string into a schema.GroupVersionKind
// Format: "group/version/kind"
func parseGVK(gvkString string) (schema.GroupVersionKind, error) {
	parts := splitAndTrim(gvkString, "/")
	if len(parts) != 3 {
		return schema.GroupVersionKind{}, fmt.Errorf("GVK must be in format 'group/version/kind', got: %s", gvkString)
	}

	group := parts[0]
	version := parts[1]
	kind := parts[2]

	// Handle core API group (empty group means "core")
	if group == "core" || group == "" {
		group = ""
	}

	return schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}, nil
}

// splitAndTrim splits a string by delimiter and trims whitespace from each part
func splitAndTrim(s, delim string) []string {
	var result []string
	for _, part := range strings.Split(s, delim) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
