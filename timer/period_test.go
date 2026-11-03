package timer

import "testing"

func TestPeriodContains(t *testing.T) {
	period := newPeriod(1024, 100)
	if period.Contains(1100) {
		t.Fatal("1100 should be outside [1000,1100)")
	}
	if !period.Contains(1000) {
		t.Fatal("1000 should be inside [1000,1100)")
	}
	if !period.Contains(1099) {
		t.Fatal("1099 should be inside [1000,1100)")
	}
}

func TestPeriodIsEarlierThan(t *testing.T) {
	period1 := newPeriod(1024, 100)
	period2 := newPeriod(1124, 100)
	if period2.IsEarlierThan(period1) {
		t.Fatal("later period should not be earlier")
	}
	if !period1.IsEarlierThan(period2) {
		t.Fatal("earlier period should be earlier")
	}
}

func TestPeriodIsLaterThan(t *testing.T) {
	period1 := newPeriod(1024, 100)
	period2 := newPeriod(1124, 100)
	if period1.IsLaterThan(period2) {
		t.Fatal("earlier period should not be later")
	}
	if !period2.IsLaterThan(period1) {
		t.Fatal("later period should be later")
	}
}

func TestPeriodGetPreviousPeriod(t *testing.T) {
	period := newPeriod(1024, 100)
	got := period.GetPreviousPeriod()
	want := newPeriod(900, 100)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPeriodGetNextPeriod(t *testing.T) {
	period := newPeriod(1024, 100)
	got := period.GetNextPeriod()
	want := newPeriod(1100, 100)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPeriodGetPeriodDifference(t *testing.T) {
	period1 := newPeriod(1024, 100)
	period2 := newPeriod(1024, 100)
	if got := period1.GetPeriodDifference(period2); got != 0 {
		t.Fatalf("got %d, want 0", got)
	}

	period2 = newPeriod(1124, 100)
	if got := period1.GetPeriodDifference(period2); got != 1 {
		t.Fatalf("got %d, want 1", got)
	}
	if got := period2.GetPeriodDifference(period1); got != 1 {
		t.Fatalf("got %d, want 1", got)
	}

	period2 = newPeriod(2024, 100)
	if got := period2.GetPeriodDifference(period1); got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
}

func TestPeriodString(t *testing.T) {
	period := newPeriod(1024, 100)
	got := period.String()
	want := "[1000-1100)"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
