package result

import (
	"errors"
	"reflect"
	"testing"
)

var errTest = errors.New("test error")

func TestResult_error(t *testing.T) {
	r := NewErrorResult[int](errTest)
	if !r.HasError() || r.HasData() {
		t.Fatal("expected error")
	}
	if got := r.Error(); !errors.Is(got, errTest) {
		t.Fatalf("Error: got %v, want %v", got, errTest)
	}
	if got := r.DataOrDefault(0); got != 0 {
		t.Fatalf("DataOrDefault: got %v, want 0", got)
	}
}

func TestResult_data(t *testing.T) {
	r := NewDataResult(100)
	if r.HasError() || !r.HasData() {
		t.Fatal("expected data")
	}
	if r.Error() != nil {
		t.Fatalf("Error: got %v, want nil", r.Error())
	}
	if got := r.Data(); got != 100 {
		t.Fatalf("Data: got %v, want 100", got)
	}
}

func TestResult_zeroValuePanics(t *testing.T) {
	var r Result[int]

	assertPanicsWith(t, func() { r.HasError() }, errNotInitialized)
	assertPanicsWith(t, func() { r.HasData() }, errNotInitialized)
	assertPanicsWith(t, func() { r.Error() }, errNotInitialized)
	assertPanicsWith(t, func() { r.DataOrDefault(0) }, errNotInitialized)
	assertPanicsWith(t, func() { r.Data() }, errNotInitialized)
}

func TestResult_someZeroValue(t *testing.T) {
	r := NewDataResult(0)
	if r.HasError() || !r.HasData() {
		t.Fatal("expected data")
	}
	if got := r.Data(); got != 0 {
		t.Fatalf("Data: got %v, want 0", got)
	}
	if got := r.DataOrDefault(7); got != 0 {
		t.Fatalf("DataOrDefault: got %v, want 0", got)
	}
}

func TestResult_dataPanics(t *testing.T) {
	assertPanicsWith(t, func() { NewErrorResult[int](errTest).Data() }, errNoData)
}

func TestResult_nilErrorPanics(t *testing.T) {
	assertPanicsWith(t, func() { NewErrorResult[int](nil) }, errNilError)
}

func TestResult_invalidStatePanics(t *testing.T) {
	r := Result[int]{err: errTest, hasData: true}

	assertPanicsWith(t, func() { r.HasError() }, errInvalidState)
	assertPanicsWith(t, func() { r.HasData() }, errInvalidState)
	assertPanicsWith(t, func() { r.Error() }, errInvalidState)
	assertPanicsWith(t, func() { r.DataOrDefault(0) }, errInvalidState)
	assertPanicsWith(t, func() { r.Data() }, errInvalidState)
}

func TestResult_types(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		r := NewDataResult("Jordan")
		if got := r.Data(); got != "Jordan" {
			t.Fatalf("got %q, want Jordan", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		want := []int64{1, 2, 3}
		r := NewDataResult(want)
		if got := r.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("map", func(t *testing.T) {
		want := map[int64]int64{1: 1}
		r := NewDataResult(want)
		if got := r.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func assertPanicsWith(t *testing.T, fn func(), want error) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("expected panic %v", want)
		}
		if got != want {
			t.Fatalf("panic: got %v, want %v", got, want)
		}
	}()

	fn()
}
