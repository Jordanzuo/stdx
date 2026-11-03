package timer

import (
	"sort"
	"sync"
)

var (
	// currPeriodTimerMutex serializes CreateTimer's AddTimer+insert with due-list
	// peel and period rebuild (snapshot + publish of globalCurrPeriod), so a
	// boundary create cannot be skipped by a rebuild or double-inserted.
	// Do not hold this lock while sending on DueTimerChannel.
	currPeriodTimerMutex sync.Mutex
	// globalCurrPeriod is the period currently being scheduled. Read and write only
	// while holding currPeriodTimerMutex (or via getGlobalCurrPeriod).
	globalCurrPeriod Period
	// currPeriodTimerList is the current period's brief timers, sorted by EndMs ascending.
	currPeriodTimerList = make([]BriefTimer, 0, 64)
)

// getGlobalCurrPeriod returns the period currently being scheduled.
func getGlobalCurrPeriod() Period {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	return globalCurrPeriod
}

// clearCurrPeriodTimerList drops every brief timer from the current due list.
func clearCurrPeriodTimerList() {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	currPeriodTimerList = currPeriodTimerList[:0]
}

// setCurrPeriodTimerListLocked sorts timerList by end time and installs it as the
// current due list. Caller must hold currPeriodTimerMutex.
func setCurrPeriodTimerListLocked(currPeriod Period, timerList []BriefTimer) {
	if len(timerList) > 0 {
		sort.Slice(timerList, func(i, j int) bool { return timerList[i].orderByEndTime(timerList[j]) })
	}
	globalCurrPeriod = currPeriod
	currPeriodTimerList = timerList
}

// initTimerUntilCurrPeriod replaces the current list with every brief timer in currPeriod
// and every earlier period, sorted by end time.
func initTimerUntilCurrPeriod(currPeriod Period) {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	setCurrPeriodTimerListLocked(currPeriod, completeTimerPtr.getBriefTimerListUntilPeriod(currPeriod))
}

// initTimerOfCurrPeriod replaces the current list with the brief timers in currPeriod only,
// sorted by end time.
func initTimerOfCurrPeriod(currPeriod Period) {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	setCurrPeriodTimerListLocked(currPeriod, completeTimerPtr.getBriefTimerListOfPeriod(currPeriod))
}

// initTimerBetweenPeriod replaces the current list with brief timers in periods
// after fromPeriod through toPeriod, sorted by end time.
// fromPeriod is excluded and toPeriod is included.
func initTimerBetweenPeriod(fromPeriod, toPeriod Period) {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	timerList := make([]BriefTimer, 0, 64)
	for _, item := range completeTimerPtr.getPeriodTimerList() {
		if !item.Period.IsLaterThan(fromPeriod) || item.Period.IsLaterThan(toPeriod) {
			continue
		}
		timerList = append(timerList, item.getBriefTimerList()...)
	}
	setCurrPeriodTimerListLocked(toPeriod, timerList)
}

// takeDueTimers returns timers that are due at nowMs, removing them from the due list.
// The channel send happens outside the lock.
func takeDueTimers(nowMs int64) []Timer {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	return getDueTimerList(nowMs)
}

// storeTimerForSchedule stores timer in completeTimerPtr and, when it ends in
// globalCurrPeriod, inserts it into the sorted due list. It serializes with
// period rebuild so the timer is not lost or double-inserted.
func storeTimerForSchedule(timer Timer) {
	currPeriodTimerMutex.Lock()
	defer currPeriodTimerMutex.Unlock()
	completeTimerPtr.addTimer(timer)
	if timer.getPeriod() == globalCurrPeriod {
		addTimerToCurrPeriodList(timer.BriefTimer)
	}
}

// addTimerToCurrPeriodList inserts briefTimer into currPeriodTimerList by end time.
// Caller must hold currPeriodTimerMutex when used on the schedule path.
func addTimerToCurrPeriodList(briefTimer BriefTimer) {
	currPeriodTimerList = append(currPeriodTimerList, briefTimer)
	if len(currPeriodTimerList) == 1 {
		return
	}

	var index int
	tail := len(currPeriodTimerList) - 1
	for index = tail - 1; index >= 0; index-- {
		item := currPeriodTimerList[index]
		if briefTimer.EndMs >= item.EndMs {
			break
		}
		currPeriodTimerList[index+1] = item
	}
	currPeriodTimerList[index+1] = briefTimer
}

// getDueTimerList returns timers whose end time is at or before nowMs, earliest first,
// and removes those brief timers from currPeriodTimerList.
// Briefs whose completeTimer entry was removed by CancelTimer are peeled without
// being sent (lazy cleanup of the due list).
// It returns nil when the current list is empty.
// Caller must hold currPeriodTimerMutex.
func getDueTimerList(nowMs int64) []Timer {
	if len(currPeriodTimerList) == 0 {
		return nil
	}

	list := make([]Timer, 0, len(currPeriodTimerList))
	deleteIndex := -1
	hasDueTimer := false
	for index, briefTimer := range currPeriodTimerList {
		if !briefTimer.IsDue(nowMs) {
			break
		}

		if timerOption := completeTimerPtr.getTimer(briefTimer); timerOption.HasData() {
			list = append(list, timerOption.Data())
		}
		deleteIndex = index
		hasDueTimer = true
	}

	if hasDueTimer {
		currPeriodTimerList = currPeriodTimerList[deleteIndex+1:]
	}
	return list
}
