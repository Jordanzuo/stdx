package timer

import (
	"sync"
	"testing"
)

func testCompleteTimer(id int64, timerType string, beginMs, endMs int64) Timer {
	return newTimer(id, timerType, beginMs, endMs, 0, 0, 0, "")
}

func TestNewCompleteTimer(t *testing.T) {
	timer := newCompleteTimer()
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
	newCompleteTimer().addTimer(testCompleteTimer(1, "type1", 0, -1))
}

func TestCompleteTimerAddTimer(t *testing.T) {
	timer := newCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)

	periodTimerList := timer.getPeriodTimerList()
	if len(periodTimerList) != 1 {
		t.Fatalf("period count = %d, want 1", len(periodTimerList))
	}
	if periodTimerList[0].getTimerCount() != 1 {
		t.Fatalf("timer count = %d, want 1", periodTimerList[0].getTimerCount())
	}
}

func TestCompleteTimerAddTimerMultiplePeriods(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	if len(timer.periodTimerMap) != 2 {
		t.Fatalf("periodTimerMap length = %d, want 2", len(timer.periodTimerMap))
	}
}

func TestCompleteTimerAddTimerConcurrentSamePeriod(t *testing.T) {
	timer := newCompleteTimer()
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

	if got := timer.getPeriodTimerCount(); got != 1 {
		t.Fatalf("period count = %d, want 1", got)
	}
	if got := len(timer.getTimerList()); got != n {
		t.Fatalf("timer count = %d, want %d", got, n)
	}
}

func TestCompleteTimerGetTimer(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))

	timerOption := timer.getTimer(NewBriefTimer(1, 1000, 2000))
	if timerOption.HasNone() {
		t.Fatal("GetTimer should return a timer")
	}
	if timerOption.Data().Id != 1 {
		t.Fatalf("timer id = %d, want 1", timerOption.Data().Id)
	}
}

func TestCompleteTimerGetTimerNotFound(t *testing.T) {
	timer := newCompleteTimer()
	if timer.getTimer(NewBriefTimer(999, 1000, 2000)).HasData() {
		t.Fatal("GetTimer should be empty for a missing timer")
	}
}

func TestCompleteTimerDeleteTimer(t *testing.T) {
	timer := newCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer2 := testCompleteTimer(2, "type1", 1000, 3000)
	timer.addTimer(timer1)
	timer.addTimer(timer2)

	timer.deleteTimer(timer1)
	if timer.getTimer(timer1.BriefTimer).HasData() {
		t.Fatal("timer should be deleted")
	}
	if timer.getPeriodTimer(timer1.getPeriod()).HasNone() {
		t.Fatal("period should remain while another timer is present")
	}
	if timer.getPeriodTimerCount() != 1 {
		t.Fatalf("period count = %d, want 1", timer.getPeriodTimerCount())
	}

	timer.deleteTimer(timer2)
	if timer.getTimer(timer2.BriefTimer).HasData() {
		t.Fatal("last timer should be deleted")
	}
	if timer.getPeriodTimer(timer2.getPeriod()).HasData() {
		t.Fatal("empty period should be removed from the map")
	}
	if timer.getPeriodTimerCount() != 0 {
		t.Fatalf("period count = %d, want 0", timer.getPeriodTimerCount())
	}
}

func TestCompleteTimerGetPeriodTimer(t *testing.T) {
	timer := newCompleteTimer()
	period := newPeriod(2000, periodIntervalMs)
	if timer.getPeriodTimer(period).HasData() {
		t.Fatal("GetPeriodTimer should be empty when the period is absent")
	}

	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)
	periodTimerOption := timer.getPeriodTimer(timer1.getPeriod())
	if periodTimerOption.HasNone() {
		t.Fatal("GetPeriodTimer should return the period after addTimer")
	}
	if periodTimerOption.Data().getTimerCount() != 1 {
		t.Fatalf("timer count = %d, want 1", periodTimerOption.Data().getTimerCount())
	}
	if periodTimerOption.Data().getTimer(timer1.Id).HasNone() {
		t.Fatal("period should contain the added timer")
	}

	timer.deletePeriodTimer(timer1.getPeriod())
	if timer.getPeriodTimer(timer1.getPeriod()).HasData() {
		t.Fatal("GetPeriodTimer should be empty after the period is deleted")
	}
}

func TestCompleteTimerDeletePeriodTimer(t *testing.T) {
	timer := newCompleteTimer()
	timer1 := testCompleteTimer(1, "type1", 1000, 2000)
	timer.addTimer(timer1)
	timer.deletePeriodTimer(timer1.getPeriod())
	if len(timer.periodTimerMap) != 0 {
		t.Fatal("period should be deleted")
	}
}

func TestCompleteTimerGetPeriodTimerCount(t *testing.T) {
	timer := newCompleteTimer()
	if got := timer.getPeriodTimerCount(); got != 0 {
		t.Fatalf("GetPeriodTimerCount() = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := timer.getPeriodTimerCount(); got != 2 {
		t.Fatalf("GetPeriodTimerCount() = %d, want 2", got)
	}
}

func TestCompleteTimerGetPeriodTimerList(t *testing.T) {
	timer := newCompleteTimer()
	if got := len(timer.getPeriodTimerList()); got != 0 {
		t.Fatalf("GetPeriodTimerList() length = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := len(timer.getPeriodTimerList()); got != 2 {
		t.Fatalf("GetPeriodTimerList() length = %d, want 2", got)
	}
}

func TestCompleteTimerGetTimerCount(t *testing.T) {
	timer := newCompleteTimer()
	if got := timer.getTimerCount(); got != 0 {
		t.Fatalf("GetTimerCount() = %d, want 0", got)
	}

	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))
	if got := timer.getTimerCount(); got != 3 {
		t.Fatalf("GetTimerCount() = %d, want 3", got)
	}
}

func TestCompleteTimerGetTimerList(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	if got := len(timer.getTimerList()); got != 2 {
		t.Fatalf("GetTimerList() length = %d, want 2", got)
	}
}

func TestCompleteTimerGetTimerListByType(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 1000, 2000))
	timer.addTimer(testCompleteTimer(2, "type1", 1000, 3000))
	timer.addTimer(testCompleteTimer(3, "type2", 0, 70000))

	if got := len(timer.getTimerListByType("type1")); got != 2 {
		t.Fatalf("type1 length = %d, want 2", got)
	}
	if got := len(timer.getTimerListByType("type2")); got != 1 {
		t.Fatalf("type2 length = %d, want 1", got)
	}
}

func TestCompleteTimerGetBriefTimerListUntilPeriod(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	timer.addTimer(testCompleteTimer(3, "type3", 0, 140000))

	list := timer.getBriefTimerListUntilPeriod(newPeriod(70000, periodIntervalMs))
	if len(list) != 2 {
		t.Fatalf("length = %d, want 2", len(list))
	}
}

func TestCompleteTimerGetBriefTimerListOfPeriod(t *testing.T) {
	timer := newCompleteTimer()
	timer.addTimer(testCompleteTimer(1, "type1", 0, 1000))
	timer.addTimer(testCompleteTimer(2, "type2", 0, 70000))
	timer.addTimer(testCompleteTimer(3, "type3", 0, 140000))

	list := timer.getBriefTimerListOfPeriod(newPeriod(70000, periodIntervalMs))
	if len(list) != 1 {
		t.Fatalf("length = %d, want 1", len(list))
	}
	if list[0].Id != 2 {
		t.Fatalf("timer id = %d, want 2", list[0].Id)
	}
}
