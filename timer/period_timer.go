package timer

import (
	"sync"

	. "github.com/Jordanzuo/stdx/option"
)

// periodTimer is the set of timers whose end times fall in one Period.
type periodTimer struct {
	// Period is the time range shared by every timer in timerMap.
	Period Period
	// timerMap stores timers in this period, keyed by id.
	timerMap map[int64]Timer
	mutex    sync.RWMutex
}

// newPeriodTimer returns an empty periodTimer for period.
func newPeriodTimer(period Period) *periodTimer {
	return &periodTimer{
		Period:   period,
		timerMap: make(map[int64]Timer),
	}
}

// getTimerCount returns how many timers this period holds.
func (this *periodTimer) getTimerCount() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return len(this.timerMap)
}

// getTimerList returns a copy of the timers in this period.
// The order is not guaranteed.
func (this *periodTimer) getTimerList() []Timer {
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

// getTimerListByType returns the timers in this period whose TimerType equals timerType.
// The order is not guaranteed.
func (this *periodTimer) getTimerListByType(timerType string) []Timer {
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

// getBriefTimerList returns the BriefTimer portion of every timer in this period.
// The order is not guaranteed.
func (this *periodTimer) getBriefTimerList() []BriefTimer {
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

// getTimer returns the timer with id, or an empty Option when it is absent.
func (this *periodTimer) getTimer(id int64) Option[Timer] {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	item, exists := this.timerMap[id]
	if !exists {
		return NewNoneOption[Timer]()
	}
	return NewDataOption(item)
}

// addTimer stores timer under its id, replacing any timer already stored with that id.
func (this *periodTimer) addTimer(timer Timer) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.timerMap[timer.Id] = timer
}

// deleteTimer removes the timer with id. A missing id leaves the map unchanged.
func (this *periodTimer) deleteTimer(id int64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.timerMap, id)
}
