package timer

import (
	"sync"
	"testing"
)

func testCompleteTimer(id int64, timerType string, beginMs, endMs int64) Timer {
	return newTimer(id, timerType, beginMs, endMs, 0, 0, 0, "")
}

func TestNewCompleteTimer(t *testing.T) {
	timer := NewCompleteTimer()
	if timer.periodTimerMap == nil {
		t.Fatal("periodTimerMap is nil")
	}
	if len(timer.periodTimerMap) != 0 {
		t.Fatalf("periodTimerMap length = %d, want 0", len(timer.periodTimerMap))
	}
}

func TestCompleteTimerAddTimerNegativeEndMs(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative EndMs should panic")
		}
	}()
	NewCompleteTimer().addTimer(testCompleteTimer(1, "type1", 0, -1))
}

func TestCompleteTimerAddTimer(t *testing.T) {
	timer := NewCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)

	periodTimerList := timer.GetPeriodTimerList()
	if len(periodTimerList) != 1 {
		t.Fatalf("period count = %d, want 1", len(periodTimerList))
	}
	if periodTimerList[0].GetTimerCount() != 1 {
		t.Fatalf("timer count = %d, want 1", periodTimerList[0].GetTimerCount())
	}
}

func TestCompleteTimerAddTimerMultiplePeriods(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	if len(timer.periodTimerMap) != 2 {
		t.Fatalf("periodTimerMap length = %d, want 2", len(timer.periodTimerMap))
	}
}

func TestCompleteTimerAddTimerConcurrentSamePeriod(t *testing.T) {
	timer := NewCompleteTimer()
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := int64(1); i <= n; i++ {
		go func(id int64) {
			defer wg.Done()
			timer.addTimer(testCompleteTimer(id, "type1", 1000, 2000))
		}(i)
	}
	wg.Wait()

	if got := timer.GetPeriodTimerCount(); got != 1 {
		t.Fatalf("period count = %d, want 1", got)
	}
	if got := len(timer.GetTimerList()); got != n {
		t.Fatalf("timer count = %d, want %d", got, n)
	}
}

func TestCompleteTimerGetTimer(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))

	timerOption := timer.GetTimer(NewBriefTimer(1, 1000, 2000))
	if timerOption.HasNone() {
		t.Fatal("GetTimer should return a timer")
	}
	if timerOption.Data().Id != 1 {
		t.Fatalf("timer id = %d, want 1", timerOption.Data().Id)
	}
}

func TestCompleteTimerGetTimerNotFound(t *testing.T) {
	timer := NewCompleteTimer()
	if timer.GetTimer(NewBriefTimer(999, 1000, 2000)).HasData() {
		t.Fatal("GetTimer should be empty for a missing timer")
	}
}

func TestCompleteTimerDeleteTimer(t *testing.T) {
	timer := NewCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer2 := testCompleteTimer(2, "type1", 1000, 3000)
	timer.addTimer(timer1)
	timer.addTimer(timer2)

	timer.deleteTimer(timer1)
	if timer.GetTimer(timer1.BriefTimer).HasData() {
		t.Fatal("timer should be deleted")
	}
	if timer.GetPeriodTimer(timer1.getPeriod()).HasNone() {
		t.Fatal("period should remain while another timer is present")
	}
	if timer.GetPeriodTimerCount() != 1 {
		t.Fatalf("period count = %d, want 1", timer.GetPeriodTimerCount())
	}

	timer.deleteTimer(timer2)
	if timer.GetTimer(timer2.BriefTimer).HasData() {
		t.Fatal("last timer should be deleted")
	}
	if timer.GetPeriodTimer(timer2.getPeriod()).HasData() {
		t.Fatal("empty period should be removed from the map")
	}
	if timer.GetPeriodTimerCount() != 0 {
		t.Fatalf("period count = %d, want 0", timer.GetPeriodTimerCount())
	}
}

func TestCompleteTimerGetPeriodTimer(t *testing.T) {
	timer := NewCompleteTimer()
	period := NewPeriod(2000, periodIntervalMs)
	if timer.GetPeriodTimer(period).HasData() {
		t.Fatal("GetPeriodTimer should be empty when the period is absent")
	}

	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)
	periodTimerOption := timer.GetPeriodTimer(timer1.getPeriod())
	if periodTimerOption.HasNone() {
		t.Fatal("GetPeriodTimer should return the period after addTimer")
	}
	if periodTimerOption.Data().GetTimerCount() != 1 {
		t.Fatalf("timer count = %d, want 1", periodTimerOption.Data().GetTimerCount())
	}
	if periodTimerOption.Data().GetTimer(timer1.Id).HasNone() {
		t.Fatal("period should contain the added timer")
	}

	timer.deletePeriodTimer(timer1.getPeriod())
	if timer.GetPeriodTimer(timer1.getPeriod()).HasData() {
		t.Fatal("GetPeriodTimer should be empty after the period is deleted")
	}
}

func TestCompleteTimerDeletePeriodTimer(t *testing.T) {
	timer := NewCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)
	timer.deletePeriodTimer(timer1.getPeriod())
	if len(timer.periodTimerMap) != 0 {
		t.Fatal("period should be deleted")
	}
}

func TestCompleteTimerGetPeriodTimerCount(t *testing.T) {
	timer := NewCompleteTimer()
	if got := timer.GetPeriodTimerCount(); got != 0 {
		t.Fatalf("GetPeriodTimerCount() = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := timer.GetPeriodTimerCount(); got != 2 {
		t.Fatalf("GetPeriodTimerCount() = %d, want 2", got)
	}
}

func TestCompleteTimerGetPeriodTimerList(t *testing.T) {
	timer := NewCompleteTimer()
	if got := len(timer.GetPeriodTimerList()); got != 0 {
		t.Fatalf("GetPeriodTimerList() length = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := len(timer.GetPeriodTimerList()); got != 2 {
		t.Fatalf("GetPeriodTimerList() length = %d, want 2", got)
	}
}

func TestCompleteTimerGetTimerCount(t *testing.T) {
	timer := NewCompleteTimer()
	if got := timer.GetTimerCount(); got != 0 {
		t.Fatalf("GetTimerCount() = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := timer.GetTimerCount(); got != 3 {
		t.Fatalf("GetTimerCount() = %d, want 3", got)
	}
}

func TestCompleteTimerGetTimerList(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	if got := len(timer.GetTimerList()); got != 2 {
		t.Fatalf("GetTimerList() length = %d, want 2", got)
	}
}

func TestCompleteTimerGetTimerListByType(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))

	if got := len(timer.GetTimerListByType("type1")); got != 2 {
		t.Fatalf("type1 length = %d, want 2", got)
	}
	if got := len(timer.GetTimerListByType("type2")); got != 1 {
		t.Fatalf("type2 length = %d, want 1", got)
	}
}

func TestCompleteTimerGetBriefTimerListUntilPeriod(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	timer.addTimer(testCompleteTimer(3, "type3", 0, 140000))

	list := timer.GetBriefTimerListUntilPeriod(NewPeriod(70000, periodIntervalMs))
	if len(list) != 2 {
		t.Fatalf("length = %d, want 2", len(list))
	}
}

func TestCompleteTimerGetBriefTimerListOfPeriod(t *testing.T) {
	timer := NewCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	timer.addTimer(testCompleteTimer(3, "type3", 0, 140000))

	list := timer.GetBriefTimerListOfPeriod(NewPeriod(70000, periodIntervalMs))
	if len(list) != 1 {
		t.Fatalf("length = %d, want 1", len(list))
	}
	if list[0].Id != 2 {
		t.Fatalf("timer id = %d, want 2", list[0].Id)
	}
}
