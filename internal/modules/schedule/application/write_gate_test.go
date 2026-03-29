package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduleWriteGateSerializesSameKey(t *testing.T) {
	gate := newScheduleWriteGate()
	start := make(chan struct{})
	var running int32
	var maxRunning int32
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := gate.WithKey(context.Background(), "same-key", func() error {
				current := atomic.AddInt32(&running, 1)
				for {
					observed := atomic.LoadInt32(&maxRunning)
					if current <= observed || atomic.CompareAndSwapInt32(&maxRunning, observed, current) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&running, -1)
				return nil
			})
			if err != nil {
				t.Errorf("gate returned error: %v", err)
			}
		}()
	}

	close(start)
	wg.Wait()

	if maxRunning != 1 {
		t.Fatalf("expected same-key work to serialize, max parallelism was %d", maxRunning)
	}
}

func TestScheduleWriteGateAllowsDifferentKeys(t *testing.T) {
	gate := newScheduleWriteGate()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondFinished := make(chan struct{})

	go func() {
		_ = gate.WithKey(context.Background(), "key-a", func() error {
			close(firstStarted)
			<-releaseFirst
			return nil
		})
	}()

	<-firstStarted

	go func() {
		_ = gate.WithKey(context.Background(), "key-b", func() error {
			close(secondFinished)
			return nil
		})
	}()

	select {
	case <-secondFinished:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected different-key work to proceed without waiting")
	}

	close(releaseFirst)
}
