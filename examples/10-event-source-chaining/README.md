# Example 10: Event Source Chaining

This example demonstrates how to use `source.Channel` to integrate external event sources with Kubernetes controllers. It shows how to receive events from outside the cluster and trigger reconciliation based on those events.

## Features

- **Channel-based Event Source**: Uses `source.Channel` to receive external events
- **Event Simulation**: Goroutine that simulates external event generation
- **Event Filtering**: Process events based on type filters
- **Event History**: Maintains history of recent events in status
- **Max Events Limit**: Configurable limit on number of events to process
- **Processing Modes**: Sequential or parallel event processing
- **Status Conditions**: Track event processing status with conditions

## Architecture

### Event Flow

```
External Source → Event Channel → Controller → Reconciliation
```

1. **External Event Generator**: A goroutine simulates external events (webhooks, messages, alerts)
2. **Event Channel**: Go channel that receives external events
3. **Type Conversion**: Converts generic events to typed events for the controller
4. **Event Handler**: Processes events and enqueues reconciliation requests
5. **Reconciliation**: Updates status with event information

### Key Components

#### ExternalHandler CRD

The custom resource that handles external events:

```yaml
apiVersion: eventsource.examples.k8s.io/v1alpha1
kind: ExternalHandler
metadata:
  name: externalhandler-sample
spec:
  eventFilter: "webhook"      # Filter events by type
  maxEvents: 50               # Max events to process
  processingMode: sequential  # sequential or parallel
```

#### Event Channel Setup

In `SetupWithManager`:

```go
// Create a typed channel for ExternalHandler events
typedChannel := make(chan event.TypedGenericEvent[*eventsourcev1alpha1.ExternalHandler], 100)

// Convert generic events to typed events
go func() {
    for evt := range r.EventChannel {
        if handler, ok := evt.Object.(*eventsourcev1alpha1.ExternalHandler); ok {
            typedChannel <- event.TypedGenericEvent[*eventsourcev1alpha1.ExternalHandler]{
                Object: handler,
            }
        }
    }
}()

// Watch the channel source
return ctrl.NewControllerManagedBy(mgr).
    For(&eventsourcev1alpha1.ExternalHandler{}).
    WatchesRawSource(
        source.Channel(
            typedChannel,
            &handler.TypedEnqueueRequestForObject[*eventsourcev1alpha1.ExternalHandler]{},
        ),
    ).
    Complete(r)
```

#### External Event Simulator

The simulator generates events every 10 seconds:

```go
func StartExternalEventSimulator(ctx context.Context, client client.Client, eventChan chan event.GenericEvent) {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                // List all ExternalHandlers
                // Generate events for active handlers
                // Send events to channel with annotations
            }
        }
    }()
}
```

## Files

- `api/v1alpha1/groupversion_info.go` - API group definition
- `api/v1alpha1/externalhandler_types.go` - ExternalHandler CRD
- `controller/externalhandler_controller.go` - Controller with channel source
- `main.go` - Main entrypoint with event simulator
- `config/crd/` - Generated CRD manifests
- `config/samples/` - Sample ExternalHandler resources

## Building

```bash
# Generate code and CRDs
controller-gen object paths="./api/..."
controller-gen crd paths="./api/..." output:crd:artifacts:config=config/crd

# Build binary
go build -o ../../bin/10-event-source-chaining main.go
```

## Running

### Prerequisites

```bash
# Apply CRDs
kubectl apply -f config/crd/

# Create sample resources
kubectl apply -f config/samples/externalhandler_sample.yaml
```

### Run the Controller

```bash
# Run locally (out-of-cluster)
../../bin/10-event-source-chaining

# Or with verbose logging
../../bin/10-event-source-chaining -zap-log-level=2
```

### Expected Behavior

1. Controller starts and begins event simulator
2. Every 10 seconds, simulator generates events for active handlers
3. Events are sent to the channel with annotations containing event details
4. Controller receives events and processes them
5. Status is updated with event information

## Testing

### Create an ExternalHandler

```bash
kubectl apply -f - <<EOF
apiVersion: eventsource.examples.k8s.io/v1alpha1
kind: ExternalHandler
metadata:
  name: test-handler
  namespace: default
spec:
  eventFilter: "webhook"
  maxEvents: 10
  processingMode: sequential
EOF
```

### Watch Status Updates

```bash
kubectl get externalhandler test-handler -o yaml -w
```

You'll see:
- `status.totalEventsProcessed` incrementing
- `status.lastProcessedEvent` with event details
- `status.recentEvents` showing last 5 events
- `status.conditions` showing processing status

### Check Events

```bash
# View recent events
kubectl get externalhandler test-handler -o jsonpath='{.status.recentEvents}' | jq .

# Check total processed
kubectl get externalhandler test-handler -o jsonpath='{.status.totalEventsProcessed}'
```

### Filter Events

Create a handler that only processes alerts:

```bash
kubectl apply -f - <<EOF
apiVersion: eventsource.examples.k8s.io/v1alpha1
kind: ExternalHandler
metadata:
  name: alert-handler
  namespace: default
spec:
  eventFilter: "alert"
  maxEvents: 5
  processingMode: sequential
EOF
```

## Use Cases

### 1. Webhook Integration

Receive webhook events from external systems (GitHub, Slack, etc.) and trigger Kubernetes actions:

```go
// HTTP webhook handler
http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
    // Parse webhook payload
    event := parseWebhook(r)

    // Create ExternalHandler with event annotations
    handler := &eventsourcev1alpha1.ExternalHandler{}
    handler.Annotations = map[string]string{
        "event.id": event.ID,
        "event.type": "webhook",
        "event.message": event.Message,
        "event.source": "github",
    }

    // Send to event channel
    eventChannel <- event.GenericEvent{Object: handler}
})
```

### 2. Message Queue Integration

Process messages from external queues (Kafka, RabbitMQ, SQS):

```go
// Kafka consumer
for msg := range kafkaConsumer.Messages() {
    handler := &eventsourcev1alpha1.ExternalHandler{}
    handler.Annotations = map[string]string{
        "event.id": string(msg.Key),
        "event.type": "message",
        "event.message": string(msg.Value),
        "event.source": "kafka",
    }

    eventChannel <- event.GenericEvent{Object: handler}
}
```

### 3. Alert Processing

Integrate with alerting systems (Prometheus Alertmanager, PagerDuty):

```go
// Alert webhook
http.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
    alerts := parseAlerts(r)

    for _, alert := range alerts {
        handler := &eventsourcev1alpha1.ExternalHandler{}
        handler.Annotations = map[string]string{
            "event.id": alert.Fingerprint,
            "event.type": "alert",
            "event.message": alert.Summary,
            "event.source": "alertmanager",
        }

        eventChannel <- event.GenericEvent{Object: handler}
    }
})
```

## Status Fields

```yaml
status:
  # Total number of events processed
  totalEventsProcessed: 15

  # Most recent event
  lastProcessedEvent:
    eventID: "evt-1728123456-15"
    eventType: "webhook"
    timestamp: "2024-10-05T10:30:45Z"
    message: "External event evt-1728123456-15 from simulator"
    source: "external-simulator"

  # Last 5 events processed
  recentEvents:
  - eventID: "evt-1728123456-15"
    eventType: "webhook"
    timestamp: "2024-10-05T10:30:45Z"
    message: "External event evt-1728123456-15 from simulator"
    source: "external-simulator"
  # ... up to 5 events

  # Active status
  isActive: true

  # Last reconcile time
  lastReconcileTime: "2024-10-05T10:30:45Z"

  # Status conditions
  conditions:
  - type: Ready
    status: "True"
    reason: Activated
    message: "Handler is ready"
  - type: Active
    status: "True"
    reason: Activated
    message: "Handler is processing events"
  - type: Processing
    status: "True"
    reason: EventProcessed
    message: "Processed event evt-1728123456-15 of type webhook"
```

## Advanced Features

### Event Filtering

Only process specific event types:

```yaml
spec:
  eventFilter: "webhook"  # Only process webhook events
```

### Processing Limits

Deactivate handler after processing a certain number of events:

```yaml
spec:
  maxEvents: 100  # Stop after 100 events
```

When limit is reached:
- `status.isActive` becomes `false`
- Condition `Active` changes to `False` with reason `Deactivated`

### Processing Modes

Choose how to process events:

```yaml
spec:
  processingMode: parallel  # Or "sequential"
```

## Key Concepts

### source.Channel

The `source.Channel` function creates an event source that reads from a Go channel:

```go
source.Channel(
    chan event.TypedGenericEvent[T],  // Typed channel
    handler.TypedEventHandler,         // Event handler
)
```

### Event Annotations

Events carry data in object annotations:

```go
handler.Annotations = map[string]string{
    "event.id": "unique-id",
    "event.type": "webhook",
    "event.message": "Event description",
    "event.source": "external-system",
}
```

### Type Conversion

Convert generic to typed events:

```go
go func() {
    for evt := range genericChannel {
        if obj, ok := evt.Object.(*MyType); ok {
            typedChannel <- event.TypedGenericEvent[*MyType]{
                Object: obj,
            }
        }
    }
}()
```

## Learning Points

1. **External Integration**: How to integrate external event sources with Kubernetes controllers
2. **Channel Sources**: Using Go channels as event sources for controllers
3. **Type Safety**: Working with typed channels and event handlers
4. **Event Processing**: Processing and tracking external events in status
5. **Goroutine Management**: Managing background goroutines in controllers
6. **Event Simulation**: Simulating external events for testing

## Next Steps

- Implement real webhook endpoint for external events
- Add message queue integration (Kafka, RabbitMQ)
- Implement event replay and retry logic
- Add metrics for event processing
- Create admission webhooks for event validation

## References

- [controller-runtime source package](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/source)
- [Event Sources Documentation](https://book.kubebuilder.io/reference/watching-resources/externally-managed.html)
- [Go Channels](https://go.dev/tour/concurrency/2)
