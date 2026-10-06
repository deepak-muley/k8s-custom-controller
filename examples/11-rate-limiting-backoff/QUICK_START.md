# Quick Start: Rate Limiting & Backoff

## Build and Run

```bash
# From the repository root
cd examples/11-rate-limiting-backoff

# Generate code and CRDs
controller-gen object:headerFile=../../hack/boilerplate.go.txt paths="./..."
controller-gen crd:crdVersions=v1 paths="./..." output:crd:artifacts:config=config/crd

# Build the binary
go build -o ../../bin/11-rate-limiting-backoff .

# Install the CRD
kubectl apply -f config/crd/

# Run the controller
../../bin/11-rate-limiting-backoff
```

## Quick Test

In another terminal:

```bash
# Apply a sample with 30% failure rate
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: test-backoff
  namespace: default
spec:
  processingTime: 2
  failureRatePercent: 30
  maxRetries: 5
  requeueAfterSeconds: 0
EOF

# Watch the resource
kubectl get ratelimited test-backoff -w

# Check detailed status
kubectl get ratelimited test-backoff -o yaml
```

## Expected Behavior

- **Processing**: Resource shows "Processing" phase during processing time
- **Success**: Phase changes to "Completed" when processing succeeds
- **Failure**: Phase changes to "RateLimited" when processing fails
- **Backoff**: Failures trigger exponential backoff (1s → 2s → 4s → 8s → 16s → max 300s)
- **Max Retries**: After max retries, phase changes to "Failed" and processing stops

## Key Status Fields

```bash
kubectl get ratelimited -o custom-columns=\
NAME:.metadata.name,\
PHASE:.status.phase,\
PROCESSING:.status.processingCount,\
FAILURES:.status.failureCount,\
BACKOFF:.status.backoffDuration,\
MESSAGE:.status.message
```

## Testing Different Scenarios

### High Failure Rate (Demonstrates Backoff)
```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: high-failure
spec:
  processingTime: 1
  failureRatePercent: 70
  maxRetries: 10
EOF
```

### Always Fail (Tests Max Retries)
```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: max-retries
spec:
  processingTime: 1
  failureRatePercent: 100
  maxRetries: 3
EOF
```

### Continuous Processing
```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: continuous
spec:
  processingTime: 3
  failureRatePercent: 0
  requeueAfterSeconds: 30
EOF
```

## Clean Up

```bash
kubectl delete ratelimited --all
kubectl delete -f config/crd/
```

## Key Concepts Demonstrated

1. **Application-Level Backoff**: Using `Result{RequeueAfter: duration}` for controlled retries
2. **Exponential Backoff**: 1s → 2s → 4s → 8s → 16s → ... → max 300s
3. **Jitter**: Random 0-25% added to prevent synchronized retries
4. **Max Retries**: Configurable limit to prevent infinite loops
5. **Status Tracking**: Phase, counts, timestamps, and backoff duration
6. **Controlled Requeue**: RequeueAfter for periodic processing

## Observing Backoff in Action

Watch the controller logs to see:
- Failure simulation (random number vs failure rate)
- Backoff calculation
- Requeue timing
- Success after retries

```bash
# Look for log entries:
# - "Failure simulation" - shows if this attempt will fail
# - "Processing failed, will retry with backoff" - shows backoff duration
# - "Processing succeeded after failures" - shows recovery
# - "Maximum retries exceeded" - shows permanent failure
```
