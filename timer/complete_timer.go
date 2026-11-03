package timer

import (
	"fmt"
	"sync"

	. "github.com/Jordanzuo/stdx/option"
)

// completeTimer stores timers grouped by the Period that contains each timer's end time.
// It is package-private; callers use CreateTimer/CancelTimer/HandleTimer and GetTimer*.
type completeTimer struct {
	// periodTimerMap stores timers by period.
	periodTimerMap map[Period]*periodTimer
	mutex          sync.RWMutex
}

// newCompleteTimer returns an empty completeTimer.
func newCompleteTimer() *completeTimer {
	return &completeTimer{
		periodTimerMap: make(map[Period]*periodTimer),
	}
}

// getPeriodTimerCount returns how many periods this collection holds.
func (this *completeTimer) getPeriodTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return len(this.periodTimerMap)
}

// getPeriodTimerList returns every period collection.
// The order is not guaranteed.
func (this *completeTimer) getPeriodTimerList() []*periodTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]*periodTimer, len(this.periodTimerMap))
	var index int
	for _, item := range this.periodTimerMap {
		list[index] = item
		index++
	}
	return list
}

// getPeriodTimer returns the periodTimer for period, or an empty Option when it is absent.
func (this *completeTimer) getPeriodTimer(period Period) Option[*periodTimer] {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	periodTimerPtr, exists := this.periodTimerMap[period]
	if !exists {
		return NewNoneOption[*periodTimer]()
	}
	return NewDataOption(periodTimerPtr)
}

// deletePeriodTimer removes the periodTimer for period.
func (this *completeTimer) deletePeriodTimer(period Period) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.periodTimerMap, period)
}

// getTimerCount returns how many timers this collection holds across every period.
func (this *completeTimer) getTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	var count int
	for _, item := range this.periodTimerMap {
		count += item.getTimerCount()
	}
	return count
}

// getTimerList returns every timer in every period.
// The order is not guaranteed.
func (this *completeTimer) getTimerList() []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	timerList := make([]Timer, 0, 1024)
	for _, item := range this.periodTimerMap {
		timerList = append(timerList, item.getTimerList()...)
	}
	return timerList
}

// getTimerListByType returns every timer whose TimerType equals timerType.
// The order is not guaranteed.
func (this *completeTimer) getTimerListByType(timerType string) []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	timerList := make([]Timer, 0, 1024)
	for _, item := range this.periodTimerMap {
		timerList = append(timerList, item.getTimerListByType(timerType)...)
	}
	return timerList
}

// getBriefTimerListUntilPeriod returns brief timers from period and every earlier period.
// The order is not guaranteed.
func (this *completeTimer) getBriefTimerListUntilPeriod(period Period) []BriefTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]BriefTimer, 0, 1024)
	for item, periodTimerPtr := range this.periodTimerMap {
		if item.IsLaterThan(period) {
			continue
		}
		list = append(list, periodTimerPtr.getBriefTimerList()...)
	}
	return list
}

// getBriefTimerListOfPeriod returns brief timers from period only.
// It returns nil when that period is absent.
func (this *completeTimer) getBriefTimerListOfPeriod(period Period) []BriefTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	periodTimerPtr, exists := this.periodTimerMap[period]
	if !exists {
		return nil
	}
	return periodTimerPtr.getBriefTimerList()
}

// getTimer returns the full timer for briefTimer, or an empty Option when it is absent.
func (this *completeTimer) getTimer(briefTimer BriefTimer) Option[Timer] {
	periodTimerOption := this.getPeriodTimer(briefTimer.getPeriod())
	if periodTimerOption.HasNone() {
		return NewNoneOption[Timer]()
	}
	return periodTimerOption.Data().getTimer(briefTimer.Id)
}

// addTimer stores timer in the period that contains its end time.
// A missing period is created. An existing id in that period is replaced.
//
// It panics when timer.EndMs is negative.
// The common path uses getPeriodTimer under a read lock. Only a missing period
// takes the write lock, and the map is checked again before creating one.
func (this *completeTimer) addTimer(timer Timer) {
	if timer.EndMs < 0 {
		panic(fmt.Sprintf("timer: EndMs %d is negative", timer.EndMs))
	}
	period := timer.getPeriod()
	periodTimerOption := this.getPeriodTimer(period)
	if periodTimerOption.HasData() {
		periodTimerOption.Data().addTimer(timer)
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	periodTimerPtr, exists := this.periodTimerMap[period]
	if !exists {
		periodTimerPtr = newPeriodTimer(period)
		this.periodTimerMap[period] = periodTimerPtr
	}
	periodTimerPtr.addTimer(timer)
}

// deleteTimer removes timer from the period that contains its end time.
// When that period has no timers left, its periodTimer is removed from the map.
// A missing period or id leaves the collection unchanged.
//
// The common path looks up the period under a read lock and deletes there.
// Only an empty period takes the write lock, and the map is checked again
// before removing it so a concurrent addTimer is not lost.
func (this *completeTimer) deleteTimer(timer Timer) {
	period := timer.getPeriod()
	periodTimerOption := this.getPeriodTimer(period)
	if periodTimerOption.HasNone() {
		return
	}
	periodTimerPtr := periodTimerOption.Data()
	periodTimerPtr.deleteTimer(timer.Id)
	if periodTimerPtr.getTimerCount() > 0 {
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	current, exists := this.periodTimerMap[period]
	if !exists || current != periodTimerPtr {
		return
	}
	if current.getTimerCount() == 0 {
		delete(this.periodTimerMap, period)
	}
}
