package utils

import "sync"

// WithRLock acquires a read lock and ensures it's released.
// This helper eliminates the need for repetitive defer mu.RUnlock() patterns.
//
// Usage:
//
//	WithRLock(&mu, func() {
//		// Read-only critical section code here
//	})
func WithRLock(mu *sync.RWMutex, fn func()) {
	mu.RLock()
	defer mu.RUnlock()
	fn()
}

// WithWLock acquires a write lock on an RWMutex and ensures it's released.
// This helper eliminates the need for repetitive defer mu.RUnlock() patterns for write locks.
//
// Usage:
//
//	WithWLock(&mu, func() {
//		// Write critical section code here
//	})
func WithWLock(mu *sync.RWMutex, fn func()) {
	mu.Lock()
	defer mu.Unlock()
	fn()
}
