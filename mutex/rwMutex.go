package mutex

import "sync"

// RWMutex provides thread-safe access to a value of type T with separate read and write locking.
//
// Use ReadData for read access; it acquires a shared read lock and allows concurrent readers.
// Lock acquires an exclusive write lock—the same exclusivity as WriteData—not a read lock.
// See the mutex package documentation for copy semantics, reference-type rules, deadlock
// avoidance, and Lock/Unlock best practices.
type RWMutex[T any] struct {
	data  T
	mutex sync.RWMutex
}

// NewRWMutex returns an RWMutex that protects data.
func NewRWMutex[T any](data T) *RWMutex[T] {
	return &RWMutex[T]{data: data}
}

// SetData replaces the protected value while holding a write lock.
func (this *RWMutex[T]) SetData(data T) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.data = data
}

// ReadData runs fn with shared read access to the protected value.
//
// Prefer ReadData over Lock when you only need to read. fn receives a copy of T; see the
// package documentation for copy and reference-type semantics. Treat fn as read-only: do
// not mutate through T (including pointer, map, or slice element updates). Multiple ReadData
// callbacks may run concurrently under a shared read lock; mutations cause data races.
// Do not retain references from fn for use after it returns.
//
// Do not call WriteData, Lock, or SetData from within fn; a read lock cannot be upgraded
// to a write lock and the call will deadlock.
func (this *RWMutex[T]) ReadData(fn func(data T)) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	fn(this.data)
}

// WriteData runs fn with exclusive write access to the protected value.
//
// fn receives a copy of T; see the package documentation for copy and reference-type
// semantics. Prefer WriteData over Lock and Unlock: the lock is released via defer,
// including when fn panics.
//
// Do not call ReadData, WriteData, Lock, or SetData from within fn on the same RWMutex;
// doing so deadlocks.
func (this *RWMutex[T]) WriteData(fn func(data T)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	fn(this.data)
}

// Lock acquires a write lock and returns a copy of the protected value.
//
// This is an exclusive write lock, not a read lock. For read-only access, use ReadData.
// When mutating value types, defer Unlock immediately after Lock, update the local copy,
// then call SetDataAfterLock before the function returns. See Mutex.Lock for the full
// Lock/Unlock pattern.
func (this *RWMutex[T]) Lock() T {
	this.mutex.Lock()
	return this.data
}

// SetDataAfterLock replaces the protected value while the write lock is already held.
//
// Call only on the goroutine that called Lock, after Lock and before Unlock. Calling
// without holding the lock causes a data race.
func (this *RWMutex[T]) SetDataAfterLock(data T) {
	this.data = data
}

// Unlock releases the write lock previously acquired by Lock on this goroutine.
//
// Calling Unlock without a matching Lock panics.
func (this *RWMutex[T]) Unlock() {
	this.mutex.Unlock()
}
