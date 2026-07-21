package tuple

import (
	"reflect"
	"testing"
)

func TestTuple2(t *testing.T) {
	tup := NewTuple2(int64(1), "a")
	if tup.First != 1 || tup.Second != "a" {
		t.Fatalf("fields: got (%v, %q)", tup.First, tup.Second)
	}
	first, second := tup.Values()
	if first != 1 || second != "a" {
		t.Fatalf("Values: got (%v, %q)", first, second)
	}
}

func TestTuple3(t *testing.T) {
	tup := NewTuple3(1, "a", true)
	assertValuesMatchFields(t, tup)
	a, b, c := tup.Values()
	if a != 1 || b != "a" || c != true {
		t.Fatalf("Values: got (%v, %q, %v)", a, b, c)
	}
}

func TestTuple4(t *testing.T) {
	tup := NewTuple4(1, 2, 3, 4)
	assertValuesMatchFields(t, tup)
}

func TestTuple5(t *testing.T) {
	tup := NewTuple5(1, 2, 3, 4, 5)
	assertValuesMatchFields(t, tup)
}

func TestTuple6(t *testing.T) {
	tup := NewTuple6(1, 2, 3, 4, 5, 6)
	assertValuesMatchFields(t, tup)
}

func TestTuple7(t *testing.T) {
	tup := NewTuple7(1, 2, 3, 4, 5, 6, 7)
	assertValuesMatchFields(t, tup)
}

func TestTuple8(t *testing.T) {
	tup := NewTuple8(1, 2, 3, 4, 5, 6, 7, 8)
	assertValuesMatchFields(t, tup)
}

func TestTuple9(t *testing.T) {
	tup := NewTuple9(1, 2, 3, 4, 5, 6, 7, 8, 9)
	assertValuesMatchFields(t, tup)
}

func TestTuple10(t *testing.T) {
	tup := NewTuple10(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	assertValuesMatchFields(t, tup)
}

func TestTuple11(t *testing.T) {
	tup := NewTuple11(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	assertValuesMatchFields(t, tup)
}

func TestTuple12(t *testing.T) {
	tup := NewTuple12(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	assertValuesMatchFields(t, tup)
}

func TestTuple13(t *testing.T) {
	tup := NewTuple13(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)
	assertValuesMatchFields(t, tup)
}

func TestTuple14(t *testing.T) {
	tup := NewTuple14(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)
	assertValuesMatchFields(t, tup)
}

func TestTuple15(t *testing.T) {
	tup := NewTuple15(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	assertValuesMatchFields(t, tup)
}

func TestTuple16(t *testing.T) {
	tup := NewTuple16(
		int64(1), "a", int8(2), int32(3), int64(4), int16(5), int64(6), int8(7),
		int32(8), float64(9), true, "z", uint8(10), uint16(11), uint32(12), float32(13),
	)
	assertValuesMatchFields(t, tup)
	a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p := tup.Values()
	if a != 1 || b != "a" || c != 2 || d != 3 || e != 4 || f != 5 || g != 6 || h != 7 ||
		i != 8 || j != 9 || k != true || l != "z" || m != 10 || n != 11 || o != 12 || p != 13 {
		t.Fatalf("unexpected Tuple16 values: %+v", tup)
	}
}

func assertValuesMatchFields(t *testing.T, tup any) {
	t.Helper()

	v := reflect.ValueOf(tup)
	values := v.MethodByName("Values").Call(nil)

	if len(values) != v.NumField() {
		t.Fatalf("Values count %d != field count %d", len(values), v.NumField())
	}

	for i := range values {
		got := values[i].Interface()
		want := v.Field(i).Interface()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("index %d: Values got %v, field got %v", i, got, want)
		}
	}
}
