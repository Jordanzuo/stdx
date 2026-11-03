package timer

import (
	"fmt"
	"sync"

	. "github.com/Jordanzuo/stdx/option"
)

// CompleteTimer stores timers grouped by the Period that contains each timer's end time.
type CompleteTimer struct {
	// periodTimerMap stores timers by period.
	periodTimerMap map[Period]*PeriodTimer
	mutex          sync.RWMutex
}

// NewCompleteTimer returns an empty CompleteTimer.
func NewCompleteTimer() *CompleteTimer {
	return &CompleteTimer{
		periodTimerMap: make(map[Period]*PeriodTimer),
	}
}

// GetPeriodTimerCount returns how many periods this collection holds.
func (this *CompleteTimer) GetPeriodTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return len(this.periodTimerMap)
}

// GetPeriodTimerList returns every period collection.
// The order is not guaranteed.
func (this *CompleteTimer) GetPeriodTimerList() []*PeriodTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]*PeriodTimer, len(this.periodTimerMap))
	var index int
	for _, item := range this.periodTimerMap {
		list[index] = item
		index++
	}
	return list
}

// GetPeriodTimer returns the PeriodTimer for period, or an empty Option when it is absent.
func (this *CompleteTimer) GetPeriodTimer(period Period) Option[*PeriodTimer] {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	periodTimerPtr, exists := this.periodTimerMap[period]
	if !exists {
		return NewNoneOption[*PeriodTimer]()
	}
	return NewDataOption(periodTimerPtr)
}

// deletePeriodTimer removes the PeriodTimer for period.
func (this *CompleteTimer) deletePeriodTimer(period Period) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.periodTimerMap, period)
}

// GetTimerCount returns how many timers this collection holds across every period.
func (this *CompleteTimer) GetTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	var count int
	for _, item := range this.periodTimerMap {
		count += item.GetTimerCount()
	}
	return count
}

// GetTimerList returns every timer in every period.
// The order is not guaranteed.
func (this *CompleteTimer) GetTimerList() []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	timerList := make([]Timer, 0, 1024)
	for _, item := range this.periodTimerMap {
		timerList = append(timerList, item.GetTimerList()...)
	}
	return timerList
}

// GetTimerListByType returns every timer whose TimerType equals timerType.
// The order is not guaranteed.
func (this *CompleteTimer) GetTimerListByType(timerType string) []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	timerList := make([]Timer, 0, 1024)
	for _, item := range this.periodTimerMap {
		timerList = append(timerList, item.GetTimerListByType(timerType)...)
	}
	return timerList
}

// GetBriefTimerListUntilPeriod returns brief timers from period and every earlier period.
// The order is not guaranteed.
func (this *CompleteTimer) GetBriefTimerListUntilPeriod(period Period) []BriefTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]BriefTimer, 0, 1024)
	for item, periodTimerPtr := range this.periodTimerMap {
		if item.IsLaterThan(period) {
			continue
		}
		list = append(list, periodTimerPtr.GetBriefTimerList()...)
	}
	return list
}

// GetBriefTimerListOfPeriod returns brief timers from period only.
// It returns nil when that period is absent.
func (this *CompleteTimer) GetBriefTimerListOfPeriod(period Period) []BriefTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	periodTimerPtr, exists := this.periodTimerMap[period]
	if !exists {
		return nil
	}
	return periodTimerPtr.GetBriefTimerList()
}

// GetTimer returns the full timer for briefTimer, or an empty Option when it is absent.
func (this *CompleteTimer) GetTimer(briefTimer BriefTimer) Option[Timer] {
	periodTimerOption := this.GetPeriodTimer(briefTimer.getPeriod())
	if periodTimerOption.HasNone() {
		return NewNoneOption[Timer]()
	}
	return periodTimerOption.Data().GetTimer(briefTimer.Id)
}

// addTimer stores timer in the period that contains its end time.
// A missing period is created. An existing id in that period is replaced.
//
// It panics when timer.EndMs is negative.
// The common path uses GetPeriodTimer under a read lock. Only a missing period
// takes the write lock, and the map is checked again before creating one.
func (this *CompleteTimer) addTimer(timer Timer) {
	if timer.EndMs < 0 {
		panic(fmt.Sprintf("timer: EndMs %d is negative", timer.EndMs))
	}
	period := timer.getPeriod()
	periodTimerOption := this.GetPeriodTimer(period)
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
// When that period has no timers left, its PeriodTimer is removed from the map.
// A missing period or id leaves the collection unchanged.
//
// The common path looks up the period under a read lock and deletes there.
// Only an empty period takes the write lock, and the map is checked again
// before removing it so a concurrent addTimer is not lost.
func (this *CompleteTimer) deleteTimer(timer Timer) {
	period := timer.getPeriod()
	periodTimerOption := this.GetPeriodTimer(period)
	if periodTimerOption.HasNone() {
		return
	}
	periodTimerPtr := periodTimerOption.Data()
	periodTimerPtr.deleteTimer(timer.Id)
	if periodTimerPtr.GetTimerCount() > 0 {
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	current, exists := this.periodTimerMap[period]
	if !exists || current != periodTimerPtr {
		return
	}
	if current.GetTimerCount() == 0 {
		delete(this.periodTimerMap, period)
	}
}
