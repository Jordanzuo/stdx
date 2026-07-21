// Package mutex provides generic mutex wrappers that protect a value with callback-based access.
//
// # Copy semantics
//
// WriteData and ReadData (on RWMutex) pass a copy of T to the callback. When T is a value
// type (struct, int, array), in-place mutations inside the callback are lost because fn
// receives a copy. Use Lock, SetDataAfterLock, and Unlock to mutate value types, or store
// a pointer as T.
//
// When T is a slice, element updates through the copied header persist because the copy
// shares the backing array. Append that triggers reallocation does not persist, because
// only the slice header copy is updated inside the callback.
//
// # Reference types
//
// When T is a pointer, map, slice, or other reference type, the callback receives a copy of
// the reference that still points at the same underlying data. In-place mutations during
// WriteData are protected by the exclusive lock. ReadData (on RWMutex) uses a shared read
// lock—multiple callbacks may run concurrently, so do not mutate through T in ReadData.
// Do not store the reference (in a variable, struct field, closure, or global) for use after
// the callback returns; concurrent access through a retained reference bypasses the mutex
// and causes data races.
//
// # Best practices
//
// Prefer WriteData over manual Lock and Unlock. WriteData uses defer to release the lock
// even when fn panics. With Lock and Unlock, always defer Unlock immediately after Lock,
// call SetDataAfterLock before the function returns, and never defer SetDataAfterLock
// after Unlock:
//
//	v := m.Lock()
//	defer m.Unlock()
//	v.Count++
//	m.SetDataAfterLock(v)
//
// SetDataAfterLock must only be called on the same goroutine that holds the lock via Lock,
// between Lock and Unlock. Calling it without holding the lock causes a data race.
//
// # Deadlock avoidance
//
// sync.Mutex is not reentrant. Do not call Lock, Unlock, WriteData, or SetData on the
// same Mutex from within a WriteData callback, or from code running while Lock is held.
// On RWMutex, do not call ReadData, WriteData, Lock, SetData, or Unlock from within a
// ReadData or WriteData callback on the same RWMutex; a read lock cannot upgrade to a
// write lock and nested locking deadlocks.
// Long-running or blocking callbacks hold the lock and stall other goroutines.
//
// Unlock releases the lock acquired by Lock. Calling Unlock without a matching Lock panics.
package mutex

import "sync"

// Mutex provides exclusive, thread-safe access to a value of type T.
type Mutex[T any] struct {
	data  T
	mutex sync.Mutex
}

// NewMutex returns a Mutex that protects data.
func NewMutex[T any](data T) *Mutex[T] {
	return &Mutex[T]{data: data}
}

// SetData replaces the protected value while holding the lock.
func (this *Mutex[T]) SetData(data T) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.data = data
}

// WriteData runs fn with exclusive access to the protected value.
//
// fn receives a copy of T; see the package documentation for copy and reference-type
// semantics. Prefer WriteData over Lock and Unlock: the lock is released via defer,
// including when fn panics.
func (this *Mutex[T]) WriteData(fn func(data T)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	fn(this.data)
}

// Lock acquires the lock and returns a copy of the protected value.
//
// Prefer WriteData when possible. When mutating value types, defer Unlock immediately
// after Lock, update the local copy, then call SetDataAfterLock before the function
// returns:
//
//	v := m.Lock()
//	defer m.Unlock()
//	v.Count++
//	m.SetDataAfterLock(v)
//
// Do not call Lock, SetData, WriteData, or SetData again on the same Mutex until Unlock
// returns on this goroutine.
func (this *Mutex[T]) Lock() T {
	this.mutex.Lock()
	return this.data
}

// SetDataAfterLock replaces the protected value while the lock is already held.
//
// Call only on the goroutine that called Lock, after Lock and before Unlock. Calling
// without holding the lock causes a data race.
func (this *Mutex[T]) SetDataAfterLock(data T) {
	this.data = data
}

// Unlock releases the lock previously acquired by Lock on this goroutine.
//
// Calling Unlock without a matching Lock panics.
func (this *Mutex[T]) Unlock() {
	this.mutex.Unlock()
}
