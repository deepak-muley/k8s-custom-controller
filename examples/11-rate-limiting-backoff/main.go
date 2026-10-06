package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	ratelimitv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/11-rate-limiting-backoff/api/v1alpha1"
	"github.com/deepak-muley/k8s-custom-controller/examples/11-rate-limiting-backoff/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(ratelimitv1alpha1.AddToScheme(scheme))
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Starting Rate Limiting & Backoff Controller",
		"version", "v1alpha1",
		"group", "ratelimit.examples.k8s.io")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ratelimited.ratelimit.examples.k8s.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup the RateLimited controller with custom rate limiter
	if err = (&controller.RateLimitedReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RateLimited")
		os.Exit(1)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Rate Limiting Configuration:",
		"workqueueBaseDelay", "1s",
		"workqueueMaxDelay", "60s",
		"perItemRateLimit", "10/second",
		"burstSize", "100",
		"applicationMaxRetries", "5 (default)",
		"applicationBaseDelay", "1s",
		"applicationMaxDelay", "300s")

	fmt.Println("\n==========================================================")
	fmt.Println("Rate Limiting & Backoff Controller Configuration")
	fmt.Println("==========================================================")
	fmt.Println("\nWorkqueue Rate Limiting:")
	fmt.Println("  - Exponential backoff: 1s base, 60s max")
	fmt.Println("  - Per-item rate limit: 10 items/second, burst 100")
	fmt.Println("\nApplication-Level Rate Limiting:")
	fmt.Println("  - Exponential backoff: 1s base, 300s max")
	fmt.Println("  - Max retries: 5 (default, configurable per resource)")
	fmt.Println("  - Jitter: 0-25% of delay")
	fmt.Println("\nReconcile Loop Configuration:")
	fmt.Println("  - MaxConcurrentReconciles: 1")
	fmt.Println("\nBackoff Pattern (with example failures):")
	fmt.Println("  Attempt 1: 1-2s")
	fmt.Println("  Attempt 2: 2-3s")
	fmt.Println("  Attempt 3: 4-5s")
	fmt.Println("  Attempt 4: 8-10s")
	fmt.Println("  Attempt 5: 16-20s")
	fmt.Println("  Max delay: 300s (5 minutes)")
	fmt.Println("\nMetrics available at:", metricsAddr)
	fmt.Println("Health probe available at:", probeAddr)
	fmt.Println("==========================================================\n")

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
