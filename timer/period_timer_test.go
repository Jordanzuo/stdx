package timer

import "testing"

func testPeriod() Period {
	return NewPeriod(1000, periodIntervalMs)
}

func TestNewPeriodTimer(t *testing.T) {
	period := testPeriod()
	periodTimer := newPeriodTimer(period)

	if periodTimer.Period != period {
		t.Fatalf("Period = %+v, want %+v", periodTimer.Period, period)
	}
	if periodTimer.timerMap == nil {
		t.Fatal("timerMap is nil")
	}
	if len(periodTimer.timerMap) != 0 {
		t.Fatalf("timerMap length = %d, want 0", len(periodTimer.timerMap))
	}
}

func TestPeriodTimerGetTimerCount(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	if got := periodTimer.GetTimerCount(); got != 0 {
		t.Fatalf("GetTimerCount() = %d, want 0", got)
	}

	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))
	if got := periodTimer.GetTimerCount(); got != 2 {
		t.Fatalf("GetTimerCount() = %d, want 2", got)
	}
}

func TestPeriodTimerGetTimerList(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))

	if got := len(periodTimer.GetTimerList()); got != 2 {
		t.Fatalf("GetTimerList() length = %d, want 2", got)
	}
}

func TestPeriodTimerGetTimerListEmpty(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	if got := len(periodTimer.GetTimerList()); got != 0 {
		t.Fatalf("GetTimerList() length = %d, want 0", got)
	}
}

func TestPeriodTimerGetTimerListByType(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(2, "type1", 1000, 3000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(3, "type2", 1000, 4000, 0, 0, 0, ""))

	if got := len(periodTimer.GetTimerListByType("type1")); got != 2 {
		t.Fatalf("type1 length = %d, want 2", got)
	}
	if got := len(periodTimer.GetTimerListByType("type2")); got != 1 {
		t.Fatalf("type2 length = %d, want 1", got)
	}
	if got := len(periodTimer.GetTimerListByType("nonexistent")); got != 0 {
		t.Fatalf("nonexistent length = %d, want 0", got)
	}
}

func TestPeriodTimerGetBriefTimerList(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))

	list := periodTimer.GetBriefTimerList()
	if len(list) != 2 {
		t.Fatalf("GetBriefTimerList() length = %d, want 2", len(list))
	}
	if list[0].Id != 1 && list[0].Id != 2 {
		t.Fatalf("brief timer id = %d, want 1 or 2", list[0].Id)
	}
}

func TestPeriodTimerGetTimer(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))

	timerOption := periodTimer.GetTimer(1)
	if timerOption.HasNone() {
		t.Fatal("GetTimer(1) should return a timer")
	}
	if timerOption.Data().Id != 1 {
		t.Fatalf("timer id = %d, want 1", timerOption.Data().Id)
	}

	if periodTimer.GetTimer(999).HasData() {
		t.Fatal("GetTimer(999) should be empty")
	}
}

func TestPeriodTimerAddTimer(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	if len(periodTimer.timerMap) != 1 {
		t.Fatalf("timerMap length = %d, want 1", len(periodTimer.timerMap))
	}

	periodTimer.addTimer(newTimer(1, "type2", 1000, 3000, 0, 0, 0, ""))
	if len(periodTimer.timerMap) != 1 {
		t.Fatalf("timerMap length = %d, want 1", len(periodTimer.timerMap))
	}
	if periodTimer.timerMap[1].TimerType != "type2" {
		t.Fatalf("timer type = %s, want type2", periodTimer.timerMap[1].TimerType)
	}
}

func TestPeriodTimerDeleteTimer(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	periodTimer.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	periodTimer.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))

	periodTimer.deleteTimer(1)
	if len(periodTimer.timerMap) != 1 {
		t.Fatalf("timerMap length = %d, want 1", len(periodTimer.timerMap))
	}
	if _, exists := periodTimer.timerMap[1]; exists {
		t.Fatal("timer 1 should be deleted")
	}
	if _, exists := periodTimer.timerMap[2]; !exists {
		t.Fatal("timer 2 should still exist")
	}

	periodTimer.deleteTimer(999)
	if len(periodTimer.timerMap) != 1 {
		t.Fatalf("timerMap length = %d, want 1", len(periodTimer.timerMap))
	}
}

func TestPeriodTimerThreadSafety(t *testing.T) {
	periodTimer := newPeriodTimer(testPeriod())
	done := make(chan bool)
	go func() {
		for i := int64(1); i <= 100; i++ {
			periodTimer.addTimer(newTimer(i, "type1", 1000, 2000, 0, 0, 0, ""))
		}
		done <- true
	}()
	go func() {
		for i := int64(101); i <= 200; i++ {
			periodTimer.addTimer(newTimer(i, "type2", 1000, 2000, 0, 0, 0, ""))
		}
		done <- true
	}()
	<-done
	<-done

	if len(periodTimer.timerMap) != 200 {
		t.Fatalf("timerMap length = %d, want 200", len(periodTimer.timerMap))
	}
}
