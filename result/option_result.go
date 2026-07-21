package result

// OptionResult holds a value of type T, an error, or neither (none).
//
// The zero value is uninitialized. Use NewErrorOptionResult, NewNoneOptionResult or NewDataOptionResult before calling any method.
type OptionResult[T any] struct {
	err     error
	none    bool
	hasData bool
	data    T
}

// NewErrorOptionResult returns an OptionResult that carries err.
//
// It panics when err is nil.
func NewErrorOptionResult[T any](err error) OptionResult[T] {
	ensureErr(err)
	return OptionResult[T]{
		err:     err,
		none:    false,
		hasData: false,
	}
}

// NewNoneOptionResult returns an empty OptionResult (no value, no error).
func NewNoneOptionResult[T any]() OptionResult[T] {
	return OptionResult[T]{
		err:     nil,
		none:    true,
		hasData: false,
	}
}

// NewDataOptionResult returns an OptionResult that carries data.
func NewDataOptionResult[T any](data T) OptionResult[T] {
	return OptionResult[T]{
		err:     nil,
		none:    false,
		hasData: true,
		data:    data,
	}
}

// HasError reports whether the OptionResult carries an error.
func (this OptionResult[T]) HasError() bool {
	this.ensureState()
	return this.err != nil
}

// Error returns the carried error, or nil when there is no error.
func (this OptionResult[T]) Error() error {
	this.ensureState()
	return this.err
}

// HasNone reports whether the OptionResult is empty.
func (this OptionResult[T]) HasNone() bool {
	this.ensureState()
	return this.none
}

// HasData reports whether the OptionResult carries a value.
func (this OptionResult[T]) HasData() bool {
	this.ensureState()
	return this.hasData
}

// Data returns the carried value. It panics when the OptionResult holds an error or is
// empty; call HasData first or use DataOrDefault for safe access.
func (this OptionResult[T]) Data() T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	panic(errNoData)
}

// DataOrDefault returns the carried value, or defaultData on error or none.
func (this OptionResult[T]) DataOrDefault(defaultData T) T {
	this.ensureState()
	if this.hasData {
		return this.data
	}
	return defaultData
}

func (this OptionResult[T]) ensureState() {
	count := 0
	if this.err != nil {
		count++
	}
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
