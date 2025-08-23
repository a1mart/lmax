# LMAX Disruptor Pattern

The LMAX Disruptor is a high-performance inter-thread messaging library
designed for low-latency and high-throughput systems. It replaces
traditional queues with a ring buffer to reduce contention and memory
allocation overhead.

## Key Concepts

-   **Ring Buffer**: A fixed-size data structure used to store events.
-   **Sequence**: Tracks positions of producers and consumers.
-   **Event Processor**: Consumes events from the ring buffer.
-   **Wait Strategy**: Determines how consumers wait for events.

## Advantages

-   **Low Latency**: Eliminates locks and reduces GC pressure.
-   **High Throughput**: Handles millions of messages per second.
-   **Deterministic Performance**: Ideal for real-time systems.

## Use Cases

-   Financial trading platforms
-   High-frequency event processing
-   Real-time analytics engines

## Core Components

  Component      Description
  -------------- ------------------------------------------
  RingBuffer     Circular buffer that stores data entries
  Sequencer      Coordinates producers and consumers
  EventHandler   Processes events
  WaitStrategy   Defines how consumers wait for events

## Implementation Overview

1.  **Initialize RingBuffer** with a power-of-two size.
2.  **Configure Sequencer** to manage sequences.
3.  **Register EventHandlers** for consuming events.
4.  **Start EventProcessors** and begin publishing events.

## Performance Tips

-   Use a power-of-two ring buffer size for bitwise optimization.
-   Pin threads to CPU cores for predictable latency.
-   Avoid object allocation inside critical paths.

## Usage
```go
ringBuffer := NewRingBuffer(10, 3)
stopCh := make(chan struct{})

// Start consumers
for i := 0; i < 3; i++ {
    go Consumer(ringBuffer, i, 2, stopCh)
}

// Produce data
for i := 0; i < 20; i += 5 {
    Producer(ringBuffer, i, i+1, i+2, i+3, i+4)
    time.Sleep(500 * time.Millisecond)
}

// Stop consumers
close(stopCh)
```

## API
`NewRingBuffer(size int, numConsumers int) *RingBuffer`
Creates a new ring buffer with the specified size and number of consumers.

`Put(data ...interface{})`
Adds data to the ring buffer in batches.

`Get(batchSize int, consumerId int) []interface{}`
Retrieves a batch of items for the specified consumer.

## Testing
### Run tests:
```bash
go test -v
```
### Run benchmarks:
```bash
go test -bench .
```

### Benchmark Summary
**Single Producer / Single Consumer**: Minimal latency and high throughput
**Multiple Producers / Multiple Consumers**: Scales efficiently with concurrency

`54.76 ns/op`

------------------------------------------------------------------------

### References

-   [LMAX Disruptor Paper](https://lmax-exchange.github.io/disruptor/)
