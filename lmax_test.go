package main

import (
	"sync"
	"testing"
	"time"
)

func TestBasicFunctionality(t *testing.T) {
	ringBuffer := NewRingBuffer(10, 2)

	// Produce data
	for i := 0; i < 10; i++ {
		Producer(ringBuffer, i)
	}

	// Consumers
	var wg sync.WaitGroup
	wg.Add(2)

	// Consumer 1
	go func() {
		defer wg.Done()
		dataBatch := ringBuffer.Get(5, 0)
		if len(dataBatch) != 5 {
			t.Errorf("Consumer 0 expected 5 items, got %d", len(dataBatch))
		}
	}()

	// Consumer 2
	go func() {
		defer wg.Done()
		dataBatch := ringBuffer.Get(5, 1)
		if len(dataBatch) != 5 {
			t.Errorf("Consumer 1 expected 5 items, got %d", len(dataBatch))
		}
	}()

	wg.Wait()
}

func TestConcurrency(t *testing.T) {
	ringBuffer := NewRingBuffer(100, 3)

	var wg sync.WaitGroup

	// Producers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				Producer(ringBuffer, id*100+j)
				time.Sleep(10 * time.Millisecond) // Simulate production time
			}
		}(i)
	}

	// Consumers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				dataBatch := ringBuffer.Get(10, id)
				if len(dataBatch) == 0 {
					return
				}
				for _, data := range dataBatch {
					t.Logf("Consumer %d consumed: %v", id, data)
				}
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkSingleProducerSingleConsumer(b *testing.B) {
	ringBuffer := NewRingBuffer(100, 1)
	stopCh := make(chan struct{})
	go Consumer(ringBuffer, 0, 10, stopCh)

	for i := 0; i < b.N; i++ {
		Producer(ringBuffer, i)
	}

	close(stopCh)
}

func BenchmarkMultipleProducersMultipleConsumers(b *testing.B) {
	ringBuffer := NewRingBuffer(1000, 3)
	stopCh := make(chan struct{})

	// Start consumers
	for i := 0; i < 3; i++ {
		go Consumer(ringBuffer, i, 10, stopCh)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Producer(ringBuffer, time.Now().UnixNano())
		}
	})

	close(stopCh)
}
