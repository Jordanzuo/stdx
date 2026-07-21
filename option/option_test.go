package option

import (
	"reflect"
	"testing"
)

func TestOption_none(t *testing.T) {
	opt := NewNoneOption[int]()
	if !opt.HasNone() || opt.HasData() {
		t.Fatal("expected none")
	}
	if got := opt.DataOrDefault(7); got != 7 {
		t.Fatalf("DataOrDefault: got %v, want 7", got)
	}
}

func TestOption_data(t *testing.T) {
	opt := NewDataOption(42)
	if opt.HasNone() || !opt.HasData() {
		t.Fatal("expected data")
	}
	if got := opt.Data(); got != 42 {
		t.Fatalf("Data: got %v, want 42", got)
	}
}

func TestOption_dataOrDefault(t *testing.T) {
	opt := NewDataOption("ok")
	if got := opt.DataOrDefault("default"); got != "ok" {
		t.Fatalf("got %q, want ok", got)
	}
}

func TestOption_zeroValuePanics(t *testing.T) {
	var opt Option[int]

	assertPanicsWith(t, func() { opt.HasNone() }, errNotInitialized)
	assertPanicsWith(t, func() { opt.HasData() }, errNotInitialized)
	assertPanicsWith(t, func() { opt.DataOrDefault(7) }, errNotInitialized)
	assertPanicsWith(t, func() { opt.Data() }, errNotInitialized)
}

func TestOption_someZeroValue(t *testing.T) {
	opt := NewDataOption(0)
	if opt.HasNone() || !opt.HasData() {
		t.Fatal("Some(0) should have data")
	}
	if got := opt.Data(); got != 0 {
		t.Fatalf("Data: got %v, want 0", got)
	}
	if got := opt.DataOrDefault(7); got != 0 {
		t.Fatalf("DataOrDefault: got %v, want 0", got)
	}
}

func TestOption_someNilPointer(t *testing.T) {
	type player struct {
		ID int64
	}

	t.Run("NewDataOption nil pointer is Some", func(t *testing.T) {
		var p *player
		opt := NewDataOption(p)
		if opt.HasNone() {
			t.Fatal("Some(nil) should not be none")
		}
		if !opt.HasData() {
			t.Fatal("Some(nil) should have data")
		}
		if got := opt.Data(); got != nil {
			t.Fatalf("Data: got %v, want nil", got)
		}
	})

	t.Run("NewNoneOption pointer type is None", func(t *testing.T) {
		opt := NewNoneOption[*player]()
		if !opt.HasNone() || opt.HasData() {
			t.Fatal("expected none")
		}
		assertPanicsWith(t, func() { opt.Data() }, errNoData)
	})

	t.Run("Some nil differs from None", func(t *testing.T) {
		someNil := NewDataOption[*player](nil)
		none := NewNoneOption[*player]()
		if someNil.HasNone() || !someNil.HasData() {
			t.Fatal("Some(nil) should have data")
		}
		if !none.HasNone() || none.HasData() {
			t.Fatal("None should not have data")
		}
	})
}

func TestOption_dataPanics(t *testing.T) {
	assertPanicsWith(t, func() { NewNoneOption[int]().Data() }, errNoData)
}

func TestOption_invalidStatePanics(t *testing.T) {
	opt := Option[int]{none: true, hasData: true}

	assertPanicsWith(t, func() { opt.HasNone() }, errInvalidState)
	assertPanicsWith(t, func() { opt.HasData() }, errInvalidState)
	assertPanicsWith(t, func() { opt.DataOrDefault(7) }, errInvalidState)
	assertPanicsWith(t, func() { opt.Data() }, errInvalidState)
}

func TestOption_hasNoneHasDataAreOpposite(t *testing.T) {
	cases := []Option[int]{
		NewNoneOption[int](),
		NewDataOption(42),
		NewDataOption(0),
	}

	for i, opt := range cases {
		if opt.HasNone() == opt.HasData() {
			t.Fatalf("case %d: HasNone and HasData must be opposite", i)
		}
	}
}

func TestOption_types(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		opt := NewDataOption("Jordan")
		if got := opt.Data(); got != "Jordan" {
			t.Fatalf("got %q, want Jordan", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		want := []int64{1, 2, 3}
		opt := NewDataOption(want)
		if got := opt.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("map", func(t *testing.T) {
		want := map[int64]int64{1: 1, 2: 2}
		opt := NewDataOption(want)
		if got := opt.Data(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		want := &struct{ ID int64 }{ID: 100}
		opt := NewDataOption(want)
		if got := opt.Data(); got != want || got.ID != 100 {
			t.Fatalf("got %+v, want %+v", got, want)
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
