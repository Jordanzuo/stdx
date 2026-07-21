// Package result provides Result and OptionResult types for modeling operation outcomes.
//
// A Result must be constructed with NewDataResult or NewErrorResult before use.
// The zero value Result[T]{} is uninitialized; calling any method on it panics.
// Use NewDataResult(0) to represent a successful zero value.
//
// An OptionResult must be constructed with NewDataOptionResult, NewErrorOptionResult,
// or NewNoneOptionResult before use. The zero value is uninitialized and panics on use.
// Use NewDataOptionResult(nil) to represent a present nil pointer, distinct from
// NewNoneOptionResult.
package result

// Result holds either a value of type T or an error.
//
// The zero value is uninitialized. Use NewErrorResult or NewDataResult before calling
// any method.
type Result[T any] struct {
	err     error
	hasData bool
	data    T
}

// NewErrorResult returns a Result that carries err.
//
// It panics when err is nil.
func NewErrorResult[T any](err error) Result[T] {
	ensureErr(err)
	return Result[T]{
		err:     err,
		hasData: false,
	}
}

// NewDataResult returns a successful Result that carries data.
func NewDataResult[T any](data T) Result[T] {
	return Result[T]{
		err:     nil,
		hasData: true,
		data:    data,
	}
}

// HasError reports whether the Result carries an error.
func (this Result[T]) HasError() bool {
	this.ensureState()
	return this.err != nil
}

// Error returns the carried error, or nil on success.
func (this Result[T]) Error() error {
	this.ensureState()
	return this.err
}

// HasData reports whether the Result carries a value.
func (this Result[T]) HasData() bool {
	this.ensureState()
	return this.hasData
}

// Data returns the carried value. It panics when the Result holds an error; call
// HasData first or use DataOrDefault for safe access.
func (this Result[T]) Data() T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	panic(errNoData)
}

// DataOrDefault returns the carried value, or defaultData when the Result holds an error.
func (this Result[T]) DataOrDefault(defaultData T) T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	return defaultData
}

func (this Result[T]) ensureState() {
	count := 0
	if this.err != nil {
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
