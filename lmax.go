package main

import (
	"fmt"
  "time"
	"sync/atomic"
	"unsafe"
)

// RingBuffer struct with advanced features
type RingBuffer struct {
	size         int
	buffer       []interface{}
	writeCursor  int64
	readCursors  []int64
	available    []int32 // Use int32 for availability tracking for atomic operations
	writeBarrier int64
	readBarrier  int64
}

// NewRingBuffer creates a new RingBuffer
func NewRingBuffer(size int, numConsumers int) *RingBuffer {
	rb := &RingBuffer{
		size:        size,
		buffer:      make([]interface{}, size),
		readCursors: make([]int64, numConsumers),
		available:   make([]int32, size), // Use int32 for availability tracking
	}
	for i := range rb.readCursors {
		rb.readCursors[i] = -1
	}
	return rb
}

// Put writes data to the ring buffer in batches
func (rb *RingBuffer) Put(data ...interface{}) {
    nextWrite := atomic.AddInt64(&rb.writeCursor, int64(len(data))) - int64(len(data))
    for _, d := range data {
        rb.buffer[nextWrite%int64(rb.size)] = d
        atomic.StoreInt32(&rb.available[nextWrite%int64(rb.size)], 1) // Mark as available
        nextWrite++
    }

    // Update write barrier to indicate the latest write operation
    //atomic.StoreInt64(&rb.writeBarrier, rb.writeCursor)
}

// Get reads data from the ring buffer in batches (lock-free)
func (rb *RingBuffer) Get(batchSize int, consumerId int) []interface{} {
	readCursor := (*int64)(unsafe.Pointer(&rb.readCursors[consumerId])) // Unsafe cast for direct memory access
	nextRead := atomic.LoadInt64(readCursor) + 1
	results := make([]interface{}, batchSize)

	index := 0 // Index for results slice
	for i := 0; i < batchSize; {
		if nextRead <= atomic.LoadInt64(&rb.writeCursor) &&
			nextRead-atomic.LoadInt64(&rb.readBarrier) < int64(rb.size) {
			if atomic.LoadInt32(&rb.available[nextRead%int64(rb.size)]) == 1 {
				results[index] = rb.buffer[nextRead%int64(rb.size)]
				atomic.StoreInt64(readCursor, nextRead)
				nextRead++
				i++
				index++
			} else {
				// Dynamic spin-wait time adjustment based on system load and expected latency
				time.Sleep(1 * time.Microsecond)
			}
		} else {
			break
		}
	}

	minReadCursor := atomic.LoadInt64(readCursor)
	for _, cursor := range rb.readCursors {
		if cursor < minReadCursor {
			minReadCursor = cursor
		}
	}
	atomic.StoreInt64(&rb.readBarrier, minReadCursor)

	return results[:index]
}

// Producer function
func Producer(rb *RingBuffer, data ...interface{}) {
	rb.Put(data...)
}

// Consumer function
func Consumer(rb *RingBuffer, id int, batchSize int, stopCh <-chan struct{}) {
	results := make([]interface{}, batchSize)
	for {
		select {
		case <-stopCh:
			fmt.Printf("Consumer %d stopping\n", id)
			return
		default:
			results = rb.Get(batchSize, id)
			for _, data := range results {
				if data != nil {
					fmt.Printf("Consumer %d consumed: %v\n", id, data)
				}
			}
		}
	}
}

func main() {
	ringBuffer := NewRingBuffer(10, 3)
	stopCh := make(chan struct{})

	// Start consumers
	for i := 0; i < 3; i++ {
		go Consumer(ringBuffer, i, 2, stopCh) // Each consumer processes batches of 2
	}

	// Produce data
	for i := 0; i < 20; i += 5 {
		Producer(ringBuffer, i, i+1, i+2, i+3, i+4)
		time.Sleep(500 * time.Millisecond) // Simulate production time
	}

	// Allow some time for consumers to process
	time.Sleep(10 * time.Second)
	close(stopCh) // Signal consumers to stop

	// Give consumers time to stop
	time.Sleep(1 * time.Second)
}
