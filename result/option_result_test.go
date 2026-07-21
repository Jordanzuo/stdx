package result

import (
	"errors"
	"reflect"
	"testing"
)

func TestOptionResult_error(t *testing.T) {
	r := NewErrorOptionResult[int](errTest)
	if !r.HasError() || r.HasData() || r.HasNone() {
		t.Fatal("expected error")
	}
	if got := r.DataOrDefault(0); got != 0 {
		t.Fatalf("DataOrDefault: got %v, want 0", got)
	}
}

func TestOptionResult_none(t *testing.T) {
	r := NewNoneOptionResult[int]()
	if !r.HasNone() || r.HasData() || r.HasError() {
		t.Fatal("expected none")
	}
	if got := r.DataOrDefault(7); got != 7 {
		t.Fatalf("DataOrDefault: got %v, want 7", got)
	}
}

func TestOptionResult_data(t *testing.T) {
	r := NewDataOptionResult(100)
	if r.HasError() || r.HasNone() || !r.HasData() {
		t.Fatal("expected data")
	}
	if r.Error() != nil {
		t.Fatalf("Error: got %v, want nil", r.Error())
	}
	if got := r.Data(); got != 100 {
		t.Fatalf("Data: got %v, want 100", got)
	}
}

func TestOptionResult_zeroValuePanics(t *testing.T) {
	var r OptionResult[int]

	assertPanicsWith(t, func() { r.HasError() }, errNotInitialized)
	assertPanicsWith(t, func() { r.HasNone() }, errNotInitialized)
	assertPanicsWith(t, func() { r.HasData() }, errNotInitialized)
	assertPanicsWith(t, func() { r.Error() }, errNotInitialized)
	assertPanicsWith(t, func() { r.DataOrDefault(0) }, errNotInitialized)
	assertPanicsWith(t, func() { r.Data() }, errNotInitialized)
}

func TestOptionResult_someZeroValue(t *testing.T) {
	r := NewDataOptionResult(0)
	if r.HasError() || r.HasNone() || !r.HasData() {
		t.Fatal("expected data")
	}
	if got := r.Data(); got != 0 {
		t.Fatalf("Data: got %v, want 0", got)
	}
	if got := r.DataOrDefault(7); got != 0 {
		t.Fatalf("DataOrDefault: got %v, want 0", got)
	}
}

func TestOptionResult_dataPanics(t *testing.T) {
	assertPanicsWith(t, func() { NewErrorOptionResult[int](errTest).Data() }, errNoData)
	assertPanicsWith(t, func() { NewNoneOptionResult[int]().Data() }, errNoData)
}

func TestOptionResult_nilErrorPanics(t *testing.T) {
	assertPanicsWith(t, func() { NewErrorOptionResult[int](nil) }, errNilError)
}

func TestOptionResult_invalidStatePanics(t *testing.T) {
	cases := []struct {
		name string
		r    OptionResult[int]
	}{
		{
			name: "err and hasData",
			r:    OptionResult[int]{err: errTest, hasData: true},
		},
		{
			name: "none and hasData",
			r:    OptionResult[int]{none: true, hasData: true},
		},
		{
			name: "err and none",
			r:    OptionResult[int]{err: errTest, none: true},
		},
		{
			name: "err none and hasData",
			r:    OptionResult[int]{err: errTest, none: true, hasData: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertPanicsWith(t, func() { tc.r.HasError() }, errInvalidState)
			assertPanicsWith(t, func() { tc.r.HasNone() }, errInvalidState)
			assertPanicsWith(t, func() { tc.r.HasData() }, errInvalidState)
			assertPanicsWith(t, func() { tc.r.Error() }, errInvalidState)
			assertPanicsWith(t, func() { tc.r.DataOrDefault(0) }, errInvalidState)
			assertPanicsWith(t, func() { tc.r.Data() }, errInvalidState)
		})
	}
}

func TestOptionResult_someNilPointer(t *testing.T) {
	type player struct {
		ID int64
	}

	t.Run("NewDataOptionResult nil pointer is Some", func(t *testing.T) {
		var p *player
		r := NewDataOptionResult(p)
		if r.HasNone() || r.HasError() {
			t.Fatal("Some(nil) should not be none or error")
		}
		if !r.HasData() {
			t.Fatal("Some(nil) should have data")
		}
		if got := r.Data(); got != nil {
			t.Fatalf("Data: got %v, want nil", got)
		}
	})

	t.Run("NewNoneOptionResult pointer type is None", func(t *testing.T) {
		r := NewNoneOptionResult[*player]()
		if !r.HasNone() || r.HasData() || r.HasError() {
			t.Fatal("expected none")
		}
		assertPanicsWith(t, func() { r.Data() }, errNoData)
	})

	t.Run("Some nil differs from None", func(t *testing.T) {
		someNil := NewDataOptionResult[*player](nil)
		none := NewNoneOptionResult[*player]()
		if someNil.HasNone() || !someNil.HasData() || someNil.HasError() {
			t.Fatal("Some(nil) should have data only")
		}
		if !none.HasNone() || none.HasData() || none.HasError() {
			t.Fatal("None should have none only")
		}
	})
}

func TestOptionResult_types(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		r := NewDataOptionResult("Jordan")
		if got := r.Data(); got != "Jordan" {
			t.Fatalf("got %q, want Jordan", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		want := []int64{1, 2, 3}
		r := NewDataOptionResult(want)
		if got := r.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("map", func(t *testing.T) {
		want := map[int64]int64{1: 1}
		r := NewDataOptionResult(want)
		if got := r.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func TestOptionResult_errorMessage(t *testing.T) {
	if got := NewErrorOptionResult[int](errTest).Error(); !errors.Is(got, errTest) {
		t.Fatalf("got %v, want %v", got, errTest)
	}
}
