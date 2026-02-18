package utils

import (
	"sync"
	"testing"
)

func TestWithRLock(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	WithRLock(&mu, func() {
		counter++
	})

	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
}

func TestWithWLock(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	WithWLock(&mu, func() {
		counter++
	})

	if counter != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter)
	}
}

func TestWithRLockConcurrent(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithRLock(&mu, func() {
				_ = counter
			})
		}()
	}
	wg.Wait()
}

func TestWithWLockConcurrent(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithWLock(&mu, func() {
				counter++
			})
		}()
	}
	wg.Wait()

	if counter != 100 {
		t.Errorf("Expected counter to be 100, got %d", counter)
	}
}

func TestWithRLockAndWLockConcurrent(t *testing.T) {
	var mu sync.RWMutex
	var counter int

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithRLock(&mu, func() {
				_ = counter
			})
		}()
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithWLock(&mu, func() {
				counter++
			})
		}()
	}
	wg.Wait()

	if counter != 50 {
		t.Errorf("Expected counter to be 50, got %d", counter)
	}
}
