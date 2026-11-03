package timer

import (
	"encoding/json"
	"testing"
)

func TestNewTimer(t *testing.T) {
	timer := newTimer(1, "testType", 1000, 2000, 123, 456, 789, "data")
	if timer.Id != 1 || timer.TimerType != "testType" || timer.IntData1 != 123 || timer.IntData2 != 456 || timer.IntData3 != 789 || timer.StringData != "data" {
		t.Fatalf("got %+v", timer)
	}
	if timer.BeginMs != 1000 || timer.EndMs != 2000 {
		t.Fatalf("bounds = %d %d, want 1000 2000", timer.BeginMs, timer.EndMs)
	}
	if !timer.IsValid() || timer.GetDurationMs() != 1000 {
		t.Fatal("embedded BriefTimer methods should apply to Timer")
	}
}

func TestTimerJSONTags(t *testing.T) {
	timer := newTimer(1, "testType", 1000, 2000, 123, 456, 789, "data")
	got, err := json.Marshal(timer)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	want := `{"id":1,"begin_ms":1000,"end_ms":2000,"timer_type":"testType","int_data1":123,"int_data2":456,"int_data3":789,"string_data":"data"}`
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
