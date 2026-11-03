package timer

import (
	"sort"
	"testing"
)

func resetCurrPeriodTimerList() {
	globalCurrPeriod = newPeriod(0, periodIntervalMs)
	currPeriodTimerList = make([]BriefTimer, 0, 64)
}

func TestInitTimerUntilCurrPeriod(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 0, 1000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 0, 70000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(3, "type3", 0, 140000, 0, 0, 0, ""))

	initTimerUntilCurrPeriod(newPeriod(70000, periodIntervalMs))
	if len(currPeriodTimerList) != 2 {
		t.Fatalf("length = %d, want 2", len(currPeriodTimerList))
	}
	if currPeriodTimerList[0].EndMs > currPeriodTimerList[1].EndMs {
		t.Fatal("list should be sorted by end time")
	}
}

func TestInitTimerBetweenPeriod(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 0, 1000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 0, 70000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(3, "type3", 0, 140000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(4, "type4", 0, 200000, 0, 0, 0, ""))

	from := newPeriod(1000, periodIntervalMs)
	to := newPeriod(140000, periodIntervalMs)
	initTimerBetweenPeriod(from, to)

	if got := getGlobalCurrPeriod(); got != to {
		t.Fatalf("globalCurrPeriod = %s, want %s", got, to)
	}
	if len(currPeriodTimerList) != 2 {
		t.Fatalf("length = %d, want 2", len(currPeriodTimerList))
	}
	if currPeriodTimerList[0].Id != 2 || currPeriodTimerList[1].Id != 3 {
		t.Fatalf("ids = %d %d, want 2 3", currPeriodTimerList[0].Id, currPeriodTimerList[1].Id)
	}
}

func TestInitTimerOfCurrPeriod(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 0, 1000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 0, 70000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(3, "type3", 0, 140000, 0, 0, 0, ""))

	initTimerOfCurrPeriod(newPeriod(70000, periodIntervalMs))
	if len(currPeriodTimerList) != 1 {
		t.Fatalf("length = %d, want 1", len(currPeriodTimerList))
	}
	if currPeriodTimerList[0].Id != 2 {
		t.Fatalf("id = %d, want 2", currPeriodTimerList[0].Id)
	}
}

func TestAddTimerToCurrPeriodList(t *testing.T) {
	resetCurrPeriodTimerList()
	addTimerToCurrPeriodList(NewBriefTimer(1, 1000, 2000))
	if len(currPeriodTimerList) != 1 {
		t.Fatalf("length = %d, want 1", len(currPeriodTimerList))
	}
}

func TestAddTimerToCurrPeriodListMaintainsSortOrder(t *testing.T) {
	resetCurrPeriodTimerList()
	addTimerToCurrPeriodList(NewBriefTimer(3, 1000, 3000))
	addTimerToCurrPeriodList(NewBriefTimer(1, 1000, 1000))
	addTimerToCurrPeriodList(NewBriefTimer(2, 1000, 2000))

	if len(currPeriodTimerList) != 3 {
		t.Fatalf("length = %d, want 3", len(currPeriodTimerList))
	}
	if currPeriodTimerList[0].EndMs != 1000 || currPeriodTimerList[1].EndMs != 2000 || currPeriodTimerList[2].EndMs != 3000 {
		t.Fatalf("end times = %d %d %d, want 1000 2000 3000", currPeriodTimerList[0].EndMs, currPeriodTimerList[1].EndMs, currPeriodTimerList[2].EndMs)
	}
}

func TestAddTimerToCurrPeriodListSingleItem(t *testing.T) {
	resetCurrPeriodTimerList()
	addTimerToCurrPeriodList(NewBriefTimer(1, 1000, 2000))
	if len(currPeriodTimerList) != 1 || currPeriodTimerList[0].Id != 1 {
		t.Fatalf("got %#v", currPeriodTimerList)
	}
}

func TestGetDueTimerList(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(3, "type3", 1000, 5000, 0, 0, 0, ""))
	initTimerUntilCurrPeriod(newPeriod(1000, periodIntervalMs))

	dueList := getDueTimerList(3500)
	if len(dueList) != 2 {
		t.Fatalf("due length = %d, want 2", len(dueList))
	}
	if len(currPeriodTimerList) != 1 || currPeriodTimerList[0].Id != 3 {
		t.Fatalf("remaining = %#v, want id 3", currPeriodTimerList)
	}
}

func TestGetDueTimerListEmpty(t *testing.T) {
	resetCurrPeriodTimerList()
	if dueList := getDueTimerList(1000); dueList != nil {
		t.Fatalf("got %#v, want nil", dueList)
	}
}

func TestGetDueTimerListNoDueTimers(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 1000, 5000, 0, 0, 0, ""))
	initTimerUntilCurrPeriod(newPeriod(1000, periodIntervalMs))

	if dueList := getDueTimerList(2000); len(dueList) != 0 {
		t.Fatalf("due length = %d, want 0", len(dueList))
	}
	if len(currPeriodTimerList) != 1 {
		t.Fatalf("remaining length = %d, want 1", len(currPeriodTimerList))
	}
}

func TestGetDueTimerListAllDue(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 1000, 3000, 0, 0, 0, ""))
	initTimerUntilCurrPeriod(newPeriod(1000, periodIntervalMs))

	if dueList := getDueTimerList(5000); len(dueList) != 2 {
		t.Fatalf("due length = %d, want 2", len(dueList))
	}
	if len(currPeriodTimerList) != 0 {
		t.Fatalf("remaining length = %d, want 0", len(currPeriodTimerList))
	}
}

func TestGetDueTimerListOrdered(t *testing.T) {
	resetCurrPeriodTimerList()
	completeTimerPtr = newCompleteTimer()
	completeTimerPtr.addTimer(newTimer(1, "type1", 1000, 1000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(2, "type2", 1000, 2000, 0, 0, 0, ""))
	completeTimerPtr.addTimer(newTimer(3, "type3", 1000, 3000, 0, 0, 0, ""))
	initTimerUntilCurrPeriod(newPeriod(1000, periodIntervalMs))

	if len(currPeriodTimerList) != 3 {
		t.Fatalf("length = %d, want 3", len(currPeriodTimerList))
	}
	if !sort.SliceIsSorted(currPeriodTimerList, func(i, j int) bool {
		return currPeriodTimerList[i].orderByEndTime(currPeriodTimerList[j])
	}) {
		t.Fatal("list should be sorted by end time")
	}
}
