package timer

import (
	"sync"

	. "github.com/Jordanzuo/stdx/option"
)

// PeriodTimer is the set of timers whose end times fall in one Period.
type PeriodTimer struct {
	// Period is the time range shared by every timer in timerMap.
	Period Period
	// timerMap stores timers in this period, keyed by id.
	timerMap map[int64]Timer
	mutex    sync.RWMutex
}

// newPeriodTimer returns an empty PeriodTimer for period.
func newPeriodTimer(period Period) *PeriodTimer {
	return &PeriodTimer{
		Period:   period,
		timerMap: make(map[int64]Timer),
	}
}

// GetTimerCount returns how many timers this period holds.
func (this *PeriodTimer) GetTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return len(this.timerMap)
}

// GetTimerList returns a copy of the timers in this period.
// The order is not guaranteed.
func (this *PeriodTimer) GetTimerList() []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]Timer, len(this.timerMap))
	var index int
	for _, item := range this.timerMap {
		list[index] = item
		index++
	}
	return list
}

// GetTimerListByType returns the timers in this period whose TimerType equals timerType.
// The order is not guaranteed.
func (this *PeriodTimer) GetTimerListByType(timerType string) []Timer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]Timer, 0, len(this.timerMap))
	for _, item := range this.timerMap {
		if item.TimerType == timerType {
			list = append(list, item)
		}
	}
	return list
}

// GetBriefTimerList returns the BriefTimer portion of every timer in this period.
// The order is not guaranteed.
func (this *PeriodTimer) GetBriefTimerList() []BriefTimer {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	list := make([]BriefTimer, len(this.timerMap))
	var index int
	for _, item := range this.timerMap {
		list[index] = item.BriefTimer
		index++
	}
	return list
}

// GetTimer returns the timer with id, or an empty Option when it is absent.
func (this *PeriodTimer) GetTimer(id int64) Option[Timer] {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	item, exists := this.timerMap[id]
	if !exists {
		return NewNoneOption[Timer]()
	}
	return NewDataOption(item)
}

// addTimer stores timer under its id, replacing any timer already stored with that id.
func (this *PeriodTimer) addTimer(timer Timer) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.timerMap[timer.Id] = timer
}

// deleteTimer removes the timer with id. A missing id leaves the map unchanged.
func (this *PeriodTimer) deleteTimer(id int64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.timerMap, id)
}
