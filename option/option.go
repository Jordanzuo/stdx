// Package option provides an optional value type for representing presence or absence
// of a value without using nil pointers.
//
// An Option must be constructed with NewNoneOption or NewDataOption before use.
// The zero value Option[T]{} is uninitialized; calling any method on it panics.
package option

import "errors"

var (
	errNoData         = errors.New("option: no data")
	errNotInitialized = errors.New("option: not initialized")
	errInvalidState   = errors.New("option: invalid state")
)

// Option is a container that may or may not hold a value of type T.
//
// The zero value is uninitialized. Use NewNoneOption or NewDataOption before calling
// any method.
type Option[T any] struct {
	none    bool
	hasData bool
	data    T
}

// NewNoneOption returns an empty Option (no value).
func NewNoneOption[T any]() Option[T] {
	return Option[T]{
		none:    true,
		hasData: false,
	}
}

// NewDataOption returns an Option that holds data.
func NewDataOption[T any](data T) Option[T] {
	return Option[T]{
		none:    false,
		hasData: true,
		data:    data,
	}
}

// HasNone reports whether the Option is empty.
func (this Option[T]) HasNone() bool {
	this.ensureState()
	return this.none
}

// HasData reports whether the Option holds a value.
func (this Option[T]) HasData() bool {
	this.ensureState()
	return this.hasData
}

// Data returns the held value. It panics when the Option is empty; call HasData first
// or use DataOrDefault for safe access.
func (this Option[T]) Data() T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	panic(errNoData)
}

// DataOrDefault returns the held value, or defaultData when the Option is empty.
func (this Option[T]) DataOrDefault(defaultData T) T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	return defaultData
}

func (this Option[T]) ensureState() {
	count := 0
	if this.none {
		count++
	}
	if this.hasData {
		count++
	}
	if count == 0 {
		panic(errNotInitialized)
	}
	if count != 1 {
		panic(errInvalidState)
	}
}
