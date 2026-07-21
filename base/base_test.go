package base

import (
	"testing"
)

func TestNew(t *testing.T) {
	if err := New("").Error(); err == nil {
		t.Fatal("expected error for empty elements")
	}

	if err := New("00").Error(); err == nil {
		t.Fatal("expected error for duplicate characters")
	}

	if r := New("01"); r.HasError() {
		t.Fatalf("unexpected error: %v", r.Error())
	}
}

func TestNewPredefinedBases(t *testing.T) {
	tests := []struct {
		name string
		base *Base
	}{
		{"base2", NewBase2()},
		{"base8", NewBase8()},
		{"base16", NewBase16(false)},
		{"base26", NewBase26(false)},
		{"base36", NewBase36(false)},
		{"base62", NewBase62()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.base == nil {
				t.Fatal("expected non-nil base")
			}
		})
	}
}

func TestTransform(t *testing.T) {
	base2 := NewBase2()
	base8 := NewBase8()
	base16 := NewBase16(false)
	base26 := NewBase26(false)
	base36 := NewBase36(false)
	base62 := NewBase62()

	tests := []struct {
		source   uint64
		base2    string
		base8    string
		base16   string
		base26   string
		base36   string
		base62   string
	}{
		{0, "0", "0", "0", "a", "0", "0"},
		{1, "1", "1", "1", "b", "1", "1"},
		{2, "10", "2", "2", "c", "2", "2"},
		{16, "10000", "20", "10", "q", "g", "g"},
		{100, "1100100", "144", "64", "dw", "2s", "1C"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assertTransform(t, base2, tt.source, tt.base2)
			assertTransform(t, base8, tt.source, tt.base8)
			assertTransform(t, base16, tt.source, tt.base16)
			assertTransform(t, base26, tt.source, tt.base26)
			assertTransform(t, base36, tt.source, tt.base36)
			assertTransform(t, base62, tt.source, tt.base62)
		})
	}
}

func TestParse(t *testing.T) {
	base2 := NewBase2()
	base8 := NewBase8()
	base16 := NewBase16(false)
	base26 := NewBase26(false)
	base36 := NewBase36(false)
	base62 := NewBase62()

	tests := []struct {
		source uint64
		base2  string
		base8  string
		base16 string
		base26 string
		base36 string
		base62 string
	}{
		{0, "0", "0", "0", "a", "0", "0"},
		{1, "1", "1", "1", "b", "1", "1"},
		{2, "10", "2", "2", "c", "2", "2"},
		{16, "10000", "20", "10", "q", "g", "g"},
		{100, "1100100", "144", "64", "dw", "2s", "1C"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assertParse(t, base2, tt.base2, tt.source)
			assertParse(t, base8, tt.base8, tt.source)
			assertParse(t, base16, tt.base16, tt.source)
			assertParse(t, base26, tt.base26, tt.source)
			assertParse(t, base36, tt.base36, tt.source)
			assertParse(t, base62, tt.base62, tt.source)
		})
	}

	if r := base62.Parse("1)"); !r.HasError() {
		t.Fatal("expected error for invalid character")
	}

	if r := base2.Parse(""); !r.HasError() {
		t.Fatal("expected error for empty source")
	}
}

func TestParseOverflow(t *testing.T) {
	base62 := NewBase62()
	max := ^uint64(0)
	maxStr := base62.Transform(max)

	if r := base62.Parse(maxStr + "0"); !r.HasError() {
		t.Fatal("expected overflow error for value above uint64 max")
	}

	if r := base62.Parse(maxStr + "1"); !r.HasError() {
		t.Fatal("expected overflow error for value above uint64 max")
	}

	long := ""
	for i := 0; i < 20; i++ {
		long += "Z"
	}
	if r := base62.Parse(long); !r.HasError() {
		t.Fatal("expected overflow error for excessively long input")
	}

	assertParse(t, base62, maxStr, max)
}

func TestInputLengthLimit(t *testing.T) {
	elements := make([]rune, maxInputLength+1)
	for i := range elements {
		elements[i] = rune(i + 32)
	}
	if r := New(string(elements)); !r.HasError() {
		t.Fatal("expected error when elements exceed max length")
	}

	base2 := NewBase2()
	valid := make([]rune, maxInputLength)
	for i := range valid {
		valid[i] = '0'
	}
	if r := base2.Parse(string(valid)); r.HasError() {
		t.Fatalf("expected len %d to be accepted: %v", maxInputLength, r.Error())
	}

	tooLong := string(append(valid, '0'))
	if r := base2.Parse(tooLong); !r.HasError() {
		t.Fatal("expected error when source exceeds max length")
	}

	huge := make([]byte, 10_000)
	for i := range huge {
		huge[i] = '0'
	}
	if r := base2.Parse(string(huge)); !r.HasError() {
		t.Fatal("expected error for huge source without full conversion")
	}

	hugeElements := make([]rune, 10_000)
	for i := range hugeElements {
		hugeElements[i] = rune(i + 32)
	}
	if r := New(string(hugeElements)); !r.HasError() {
		t.Fatal("expected error for huge elements without full conversion")
	}
}

func TestTransformUppercase(t *testing.T) {
	base16 := NewBase16(true)
	base26 := NewBase26(true)
	base36 := NewBase36(true)

	assertTransform(t, base16, 255, "FF")
	assertTransform(t, base26, 0, "A")
	assertTransform(t, base26, 16, "Q")
	assertTransform(t, base36, 16, "G")
	assertTransform(t, base36, 35, "Z")
}

func TestParseUppercase(t *testing.T) {
	base16 := NewBase16(true)
	base26 := NewBase26(true)
	base36 := NewBase36(true)

	assertParse(t, base16, "FF", 255)
	assertParse(t, base26, "Q", 16)
	assertParse(t, base36, "G", 16)
	assertParse(t, base36, "Z", 35)
}

func TestNonASCIIBase(t *testing.T) {
	b := New("あい").Data()

	assertTransform(t, b, 0, "あ")
	assertTransform(t, b, 1, "い")
	assertTransform(t, b, 2, "いあ")
	assertTransform(t, b, 3, "いい")

	assertParse(t, b, "あ", 0)
	assertParse(t, b, "い", 1)
	assertParse(t, b, "いあ", 2)
	assertParse(t, b, "いい", 3)
}

func assertTransform(t *testing.T, b *Base, source uint64, want string) {
	t.Helper()
	if got := b.Transform(source); got != want {
		t.Fatalf("Transform(%d): got %q, want %q", source, got, want)
	}
}

func assertParse(t *testing.T, b *Base, source string, want uint64) {
	t.Helper()
	r := b.Parse(source)
	if r.HasError() {
		t.Fatalf("Parse(%q): unexpected error: %v", source, r.Error())
	}
	if got := r.Data(); got != want {
		t.Fatalf("Parse(%q): got %d, want %d", source, got, want)
	}
}
