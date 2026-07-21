package math

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/Jordanzuo/stdx/result"
)

func parseUintBound(s string) result.Result[uint] {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return result.NewErrorResult[uint](err)
	}
	return result.NewDataResult(uint(v))
}

func mustInterval(t *testing.T, lower, upper uint, intervalType IntervalType) Interval[uint] {
	t.Helper()

	r := BuildInterval(lower, upper, intervalType)
	if r.HasError() {
		t.Fatalf("BuildInterval(%d, %d, %v): unexpected error: %v", lower, upper, intervalType, r.Error())
	}
	return r.Data()
}

func TestBuildInterval(t *testing.T) {
	tests := []struct {
		name         string
		lower        uint
		upper        uint
		intervalType IntervalType
		wantLower    uint
		wantUpper    uint
	}{
		{"inclusive inclusive", 1, 5, InclusiveInclusive, 1, 5},
		{"inclusive exclusive", 2, 6, InclusiveExclusive, 2, 6},
		{"exclusive inclusive", 3, 7, ExclusiveInclusive, 3, 7},
		{"exclusive exclusive", 4, 8, ExclusiveExclusive, 4, 8},
		{"equal bounds", 10, 10, InclusiveInclusive, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustInterval(t, tt.lower, tt.upper, tt.intervalType)
			if got.Lower != tt.wantLower || got.Upper != tt.wantUpper || got.Type != tt.intervalType {
				t.Fatalf("got %+v, want {Lower:%d Upper:%d Type:%v}", got, tt.wantLower, tt.wantUpper, tt.intervalType)
			}
		})
	}

	r := BuildInterval[uint](5, 1, InclusiveInclusive)
	if !r.HasError() {
		t.Fatal("expected error when lower exceeds upper")
	}
	if !strings.Contains(r.Error().Error(), "lower bound exceeds upper bound") {
		t.Fatalf("unexpected error: %v", r.Error())
	}

	r = BuildInterval[uint](1, 5, IntervalType(0))
	if !r.HasError() {
		t.Fatal("expected error for invalid interval type")
	}
	if !strings.Contains(r.Error().Error(), "invalid interval type") {
		t.Fatalf("unexpected error: %v", r.Error())
	}

	for _, intervalType := range []IntervalType{InclusiveExclusive, ExclusiveInclusive, ExclusiveExclusive} {
		r = BuildInterval[uint](3, 3, intervalType)
		if !r.HasError() {
			t.Fatalf("expected error for empty interval type %v", intervalType)
		}
		if !strings.Contains(r.Error().Error(), "empty interval") {
			t.Fatalf("unexpected error: %v", r.Error())
		}
	}

	r = BuildInterval[uint](0, ^uint(0), InclusiveInclusive)
	if !r.HasError() {
		t.Fatal("expected error for full-range inclusive interval")
	}
	if !strings.Contains(r.Error().Error(), "interval length overflow") {
		t.Fatalf("unexpected error: %v", r.Error())
	}

	r8 := BuildInterval[uint8](0, ^uint8(0), InclusiveInclusive)
	if !r8.HasError() {
		t.Fatal("expected error for full-range inclusive uint8 interval")
	}

	r = BuildInterval[uint](0, ^uint(0), InclusiveExclusive)
	if r.HasError() {
		t.Fatalf("unexpected error for half-open full range: %v", r.Error())
	}
	if got := r.Data().Length(); got != ^uint(0) {
		t.Fatalf("half-open full range Length(): got %d, want %d", got, ^uint(0))
	}
}

func TestBuildIntervalFromString(t *testing.T) {
	tests := []struct {
		input        string
		intervalType IntervalType
		wantLower    uint
		wantUpper    uint
	}{
		{"[1,5]", InclusiveInclusive, 1, 5},
		{"[2,6)", InclusiveExclusive, 2, 6},
		{"(3,7]", ExclusiveInclusive, 3, 7},
		{"(4,8)", ExclusiveExclusive, 4, 8},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := BuildIntervalFromString(tt.input, parseUintBound)
			if r.HasError() {
				t.Fatalf("unexpected error: %v", r.Error())
			}
			got := r.Data()
			if got.Lower != tt.wantLower || got.Upper != tt.wantUpper || got.Type != tt.intervalType {
				t.Fatalf("got %+v, want {Lower:%d Upper:%d Type:%v}", got, tt.wantLower, tt.wantUpper, tt.intervalType)
			}
		})
	}

	t.Run("empty input", func(t *testing.T) {
		r := BuildIntervalFromString("", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		r := BuildIntervalFromString("{1,5}", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error for invalid format")
		}
	})

	t.Run("invalid lower bound", func(t *testing.T) {
		r := BuildIntervalFromString("[x,5]", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error for invalid lower bound")
		}
	})

	t.Run("invalid upper bound", func(t *testing.T) {
		r := BuildIntervalFromString("[1,x]", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error for invalid upper bound")
		}
	})

	t.Run("lower exceeds upper", func(t *testing.T) {
		r := BuildIntervalFromString("[5,1]", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error when lower exceeds upper")
		}
		if !strings.Contains(r.Error().Error(), "lower bound exceeds upper bound") {
			t.Fatalf("unexpected error: %v", r.Error())
		}
	})

	t.Run("signed lower bound syntax fails conversion", func(t *testing.T) {
		r := BuildIntervalFromString("[-1,5]", parseUintBound)
		if !r.HasError() {
			t.Fatal("expected error when lower bound is not unsigned")
		}
	})

	t.Run("convert func error preserved", func(t *testing.T) {
		want := errors.New("convert failed")
		r := BuildIntervalFromString("[1,5]", func(string) result.Result[uint] {
			return result.NewErrorResult[uint](want)
		})
		if !r.HasError() {
			t.Fatal("expected convert error")
		}
		if !errors.Is(r.Error(), want) {
			t.Fatalf("got error %v, want %v", r.Error(), want)
		}
	})
}

func TestIntervalString(t *testing.T) {
	tests := []struct {
		interval Interval[uint]
		want     string
	}{
		{Interval[uint]{Lower: 1, Upper: 5, Type: InclusiveInclusive}, "[1,5]"},
		{Interval[uint]{Lower: 2, Upper: 6, Type: InclusiveExclusive}, "[2,6)"},
		{Interval[uint]{Lower: 3, Upper: 7, Type: ExclusiveInclusive}, "(3,7]"},
		{Interval[uint]{Lower: 4, Upper: 8, Type: ExclusiveExclusive}, "(4,8)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.interval.String(); got != tt.want {
				t.Fatalf("String(): got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIntervalContains(t *testing.T) {
	inclusive := mustInterval(t, 1, 5, InclusiveInclusive)
	halfOpenLower := mustInterval(t, 1, 5, InclusiveExclusive)
	halfOpenUpper := mustInterval(t, 1, 5, ExclusiveInclusive)
	exclusive := mustInterval(t, 1, 5, ExclusiveExclusive)

	tests := []struct {
		name     string
		interval Interval[uint]
		value    uint
		want     bool
	}{
		{"inclusive inside", inclusive, 3, true},
		{"inclusive lower", inclusive, 1, true},
		{"inclusive upper", inclusive, 5, true},
		{"inclusive below", inclusive, 0, false},
		{"inclusive above", inclusive, 6, false},
		{"half open lower inside", halfOpenLower, 4, true},
		{"half open lower upper excluded", halfOpenLower, 5, false},
		{"half open upper inside", halfOpenUpper, 2, true},
		{"half open upper lower excluded", halfOpenUpper, 1, false},
		{"exclusive inside", exclusive, 3, true},
		{"exclusive lower excluded", exclusive, 1, false},
		{"exclusive upper excluded", exclusive, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Contains(tt.value); got != tt.want {
				t.Fatalf("Contains(%d): got %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIntervalLength(t *testing.T) {
	tests := []struct {
		name     string
		interval Interval[uint]
		want     uint
	}{
		{"inclusive inclusive", mustInterval(t, 1, 5, InclusiveInclusive), 5},
		{"inclusive exclusive", mustInterval(t, 1, 5, InclusiveExclusive), 4},
		{"exclusive inclusive", mustInterval(t, 1, 5, ExclusiveInclusive), 4},
		{"exclusive exclusive", mustInterval(t, 1, 5, ExclusiveExclusive), 3},
		{"single inclusive", mustInterval(t, 3, 3, InclusiveInclusive), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Length(); got != tt.want {
				t.Fatalf("Length(): got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIntervalDistanceFromBeginning(t *testing.T) {
	interval := mustInterval(t, 10, 20, InclusiveInclusive)

	if got := interval.DistanceFromBeginning(15); got != 5 {
		t.Fatalf("DistanceFromBeginning(15): got %d, want 5", got)
	}
	if got := interval.DistanceFromBeginning(5); got != 0 {
		t.Fatalf("DistanceFromBeginning(5): got %d, want 0", got)
	}
	if got := interval.DistanceFromBeginning(25); got != 0 {
		t.Fatalf("DistanceFromBeginning(25): got %d, want 0", got)
	}
}

func TestStringRoundTrip(t *testing.T) {
	interval := mustInterval(t, 12, 34, InclusiveExclusive)
	r := BuildIntervalFromString(interval.String(), parseUintBound)
	if r.HasError() {
		t.Fatalf("unexpected error: %v", r.Error())
	}
	got := r.Data()
	if got != interval {
		t.Fatalf("round trip: got %+v, want %+v", got, interval)
	}
}
