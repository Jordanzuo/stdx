package timer

import (
	"testing"
	"time"
)

func TestNewBriefTimer(t *testing.T) {
	timer := NewBriefTimer(1, 1000, 2000)
	if timer.Id != 1 || timer.BeginMs != 1000 || timer.EndMs != 2000 {
		t.Fatalf("got %+v, want id 1 begin 1000 end 2000", timer)
	}
}

func TestBriefTimerIsValid(t *testing.T) {
	if !NewBriefTimer(1, 1000, 2000).IsValid() {
		t.Fatal("positive id should be valid")
	}
	if NewBriefTimer(0, 1000, 2000).IsValid() {
		t.Fatal("zero id should be invalid")
	}
	if NewBriefTimer(-1, 1000, 2000).IsValid() {
		t.Fatal("negative id should be invalid")
	}
}

func TestBriefTimerIsDue(t *testing.T) {
	timer := NewBriefTimer(1, 1000, 2000)
	if timer.IsDue(1500) {
		t.Fatal("1500 should not be due")
	}
	if !timer.IsDue(2000) {
		t.Fatal("2000 should be due")
	}
	if !timer.IsDue(2500) {
		t.Fatal("2500 should be due")
	}
}

func TestBriefTimerGetDuration(t *testing.T) {
	timer := NewBriefTimer(1, 0, 5000)
	if got := timer.GetDurationMs(); got != 5000 {
		t.Fatalf("GetDurationMs() = %d, want 5000", got)
	}
	if got := timer.GetDurationSec(); got != 5 {
		t.Fatalf("GetDurationSec() = %d, want 5", got)
	}
}

func TestBriefTimerGetElapsedMs(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-500, now+500)
	got := timer.GetElapsedMs()
	if got < 400 || got > 700 {
		t.Fatalf("GetElapsedMs() = %d, want about 500", got)
	}
}

func TestBriefTimerGetElapsedMsAfterEnd(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-2000, now-1000)
	if got := timer.GetElapsedMs(); got != 1000 {
		t.Fatalf("GetElapsedMs() = %d, want 1000", got)
	}
}

func TestBriefTimerGetElapsedMsBeforeBegin(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now+1000, now+2000)
	if got := timer.GetElapsedMs(); got != 0 {
		t.Fatalf("GetElapsedMs() = %d, want 0", got)
	}
	if got := timer.GetElapsedSec(); got != 0 {
		t.Fatalf("GetElapsedSec() = %d, want 0", got)
	}
}

func TestBriefTimerGetElapsedMsInvalid(t *testing.T) {
	if got := NewBriefTimer(0, 1000, 2000).GetElapsedMs(); got != 0 {
		t.Fatalf("GetElapsedMs() = %d, want 0", got)
	}
}

func TestBriefTimerGetElapsedSec(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-5500, now+4500)
	got := timer.GetElapsedSec()
	if got < 4 || got > 6 {
		t.Fatalf("GetElapsedSec() = %d, want about 5", got)
	}
}

func TestBriefTimerGetRemainMs(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-500, now+500)
	got := timer.GetRemainMs()
	if got < 300 || got > 600 {
		t.Fatalf("GetRemainMs() = %d, want about 500", got)
	}
}

func TestBriefTimerGetRemainMsAfterEnd(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-2000, now-1000)
	if got := timer.GetRemainMs(); got != 0 {
		t.Fatalf("GetRemainMs() = %d, want 0", got)
	}
}

func TestBriefTimerGetRemainMsInvalid(t *testing.T) {
	if got := NewBriefTimer(0, 1000, 2000).GetRemainMs(); got != 0 {
		t.Fatalf("GetRemainMs() = %d, want 0", got)
	}
}

func TestBriefTimerGetRemainSec(t *testing.T) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-2000, now+3000)
	got := timer.GetRemainSec()
	if got < 2 || got > 4 {
		t.Fatalf("GetRemainSec() = %d, want about 3", got)
	}
}

func TestBriefTimerOrderByEndTime(t *testing.T) {
	timer1 := NewBriefTimer(1, 1000, 2000)
	timer2 := NewBriefTimer(2, 1000, 3000)
	if !timer1.orderByEndTime(timer2) {
		t.Fatal("earlier end should order first")
	}
	if timer2.orderByEndTime(timer1) {
		t.Fatal("later end should not order first")
	}
}

func TestBriefTimerGetPeriod(t *testing.T) {
	timer := NewBriefTimer(1, 0, 90000)
	got := timer.getPeriod()
	want := newPeriod(90000, periodIntervalMs)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

