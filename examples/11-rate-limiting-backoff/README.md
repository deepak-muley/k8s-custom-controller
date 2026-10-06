# Example 11: Rate Limiting & Backoff

This example demonstrates advanced rate limiting and backoff strategies in Kubernetes controllers, including:

- Custom workqueue rate limiting configuration
- Exponential backoff with jitter
- Per-item and global rate limiting
- Controlled retries with RequeueAfter
- Application-level vs workqueue-level rate limiting

## Components

### API Types (api/v1alpha1/)

- **RateLimited**: A custom resource that simulates processing with configurable failure rates
  - `processingTime`: Time to simulate processing (0-300 seconds)
  - `failureRate`: Probability of failure (0.0-1.0)
  - `maxRetries`: Maximum retry attempts before giving up
  - `requeueAfterSeconds`: Time to wait before requeuing after success

### Controller (controller/)

**RateLimitedReconciler** demonstrates:

1. **Workqueue Rate Limiting**:
   - Exponential backoff rate limiter (1s base, 60s max)
   - Per-item rate limiter (10 items/second, burst of 100)
   - Combined using `NewMaxOfRateLimiter`

2. **Application-Level Backoff**:
   - Custom exponential backoff calculation
   - Jitter (0-25% of delay) to prevent thundering herd
   - Max delay cap (300 seconds)

3. **Retry Logic**:
   - Tracks failure count and processing count
   - Returns `Result{RequeueAfter: duration}` for controlled retries
   - Gives up after max retries exceeded

4. **Status Tracking**:
   - Phase: Pending, Processing, Completed, Failed, RateLimited
   - Processing and failure counts
   - Timestamps for last processed and last failure
   - Current backoff duration

## Rate Limiting Concepts

### Workqueue-Level Rate Limiting

Controller-runtime uses workqueues that support rate limiting at the queue level:

```go
rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(
    1*time.Second,  // Base delay
    60*time.Second, // Max delay
)

// Combine with per-item rate limiter
rateLimiter = workqueue.NewMaxOfRateLimiter(
    rateLimiter,
    &workqueue.BucketRateLimiter{Limiter: newRateLimiter(10, 100)},
)
```

**Key Points**:
- Applied when items are added back to the queue after failure
- Prevents overwhelming the API server
- Separate from application logic

### Application-Level Backoff

Implemented in reconcile logic using `Result{RequeueAfter: duration}`:

```go
backoffSeconds := calculateBackoff(failureCount)
return ctrl.Result{RequeueAfter: time.Duration(backoffSeconds) * time.Second}, nil
```

**Backoff Pattern** (with jitter):
- Attempt 1: 1-2 seconds
- Attempt 2: 2-3 seconds
- Attempt 3: 4-5 seconds
- Attempt 4: 8-10 seconds
- Attempt 5: 16-20 seconds
- Max: 300 seconds (5 minutes)

**Key Points**:
- Controls retry timing based on business logic
- Can incorporate resource-specific state
- Adds jitter to prevent synchronized retries

## Building and Running

### Generate CRDs and DeepCopy

```bash
cd examples/11-rate-limiting-backoff

# Generate DeepCopy methods
controller-gen object:headerFile=../../hack/boilerplate.go.txt paths="./..."

# Generate CRDs
controller-gen crd:crdVersions=v1 paths="./..." output:crd:artifacts:config=config/crd
```

### Build the Controller

```bash
# Build binary
go build -o ../../bin/11-rate-limiting-backoff .

# Or use the Makefile from the root directory
cd ../..
make build-11
```

### Install CRDs

```bash
kubectl apply -f examples/11-rate-limiting-backoff/config/crd/
```

### Run the Controller

```bash
./bin/11-rate-limiting-backoff
```

## Testing Examples

### Example 1: Low Failure Rate

```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: low-failure
  namespace: default
spec:
  processingTime: 2
  failureRate: 0.3  # 30% chance of failure
  maxRetries: 5
EOF
```

Watch the processing:
```bash
kubectl get ratelimited low-failure -w
```

Check status:
```bash
kubectl get ratelimited low-failure -o yaml
```

### Example 2: High Failure Rate (Backoff Demonstration)

```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: high-failure
  namespace: default
spec:
  processingTime: 1
  failureRate: 0.7  # 70% chance of failure
  maxRetries: 10
EOF
```

This will demonstrate exponential backoff as failures occur.

### Example 3: Continuous Processing

```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: continuous
  namespace: default
spec:
  processingTime: 3
  failureRate: 0.0
  requeueAfterSeconds: 30  # Reprocess every 30 seconds
EOF
```

### Example 4: Always Fail (Max Retries)

```bash
kubectl apply -f - <<EOF
apiVersion: ratelimit.examples.k8s.io/v1alpha1
kind: RateLimited
metadata:
  name: max-retries-test
  namespace: default
spec:
  processingTime: 1
  failureRate: 1.0  # Always fail
  maxRetries: 3
EOF
```

This will fail 3 times and then give up.

### Apply All Examples

```bash
kubectl apply -f examples/11-rate-limiting-backoff/config/samples/
```

## Observing Rate Limiting

### Watch All Resources

```bash
kubectl get ratelimited -w
```

### Check Detailed Status

```bash
kubectl get ratelimited -o custom-columns=\
NAME:.metadata.name,\
PHASE:.status.phase,\
PROCESSING:.status.processingCount,\
FAILURES:.status.failureCount,\
BACKOFF:.status.backoffDuration,\
MESSAGE:.status.message
```

### View Controller Logs

The controller logs show:
- Processing attempts
- Failure simulation results
- Backoff calculations
- Requeue decisions

```bash
# Look for log entries like:
# - "Starting reconciliation"
# - "Simulating processing time"
# - "Failure simulation"
# - "Processing failed, will retry with backoff"
# - "Processing succeeded after failures"
```

## Key Learnings

1. **Two-Level Rate Limiting**:
   - Workqueue level: Protects infrastructure
   - Application level: Implements business logic

2. **RequeueAfter vs Return Error**:
   - `RequeueAfter`: Controlled retry timing
   - `Return error`: Uses workqueue rate limiter

3. **Exponential Backoff Benefits**:
   - Reduces load during failures
   - Increases success probability over time
   - Prevents overwhelming external services

4. **Jitter Importance**:
   - Prevents synchronized retries
   - Spreads load over time
   - Reduces thundering herd effect

5. **Max Retries**:
   - Prevents infinite retry loops
   - Allows manual intervention
   - Sets clear failure boundaries

## Cleanup

```bash
kubectl delete -f examples/11-rate-limiting-backoff/config/samples/
kubectl delete -f examples/11-rate-limiting-backoff/config/crd/
```

## Real-World Applications

This pattern is useful for:

1. **External API Integration**:
   - Respect rate limits of external services
   - Handle temporary failures gracefully

2. **Resource-Intensive Operations**:
   - Throttle CPU/memory intensive tasks
   - Prevent cluster overload

3. **Cascading Failures**:
   - Give downstream services time to recover
   - Implement circuit breaker patterns

4. **Cost Control**:
   - Limit expensive operations (cloud API calls)
   - Control resource consumption

5. **Batch Processing**:
   - Process items at controlled rate
   - Prevent overwhelming data stores
