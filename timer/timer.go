package timer

import (
	"fmt"
	"math"
	"time"

	. "github.com/Jordanzuo/stdx/mutex"
	. "github.com/Jordanzuo/stdx/option"
)

var (
	// completeTimerPtr is the package collection of timers grouped by period.
	completeTimerPtr = newCompleteTimer()
	// DueTimerChannel receives timers that schedule finds due.
	DueTimerChannel = make(chan Timer, dueTimerChannelCapacity)

	// lifecycleMutex protects the Init/Start/Stop state machine.
	// CreateTimer and CancelTimer hold a read lock for the whole call so Stop cannot
	// move to stopped until they finish. Init/Start/Stop take a write lock.
	lifecycleMutex = NewRWMutex(idle)
	closingChannel = make(chan struct{}, 1)
	closedChannel  = make(chan struct{})
)

// Init adds timerList to completeTimerPtr and sets the last issued id to the
// maximum Id in timerList (0 when the list is empty). It then fills the sorted
// current-period list with every timer in the current minute and every earlier
// period, and sets globalCurrPeriod to that minute.
//
// Init, Start, and Stop must be called once in that order (idle → inited → running → stopped).
// Init panics unless the lifecycle is idle. After Stop, the lifecycle stays stopped;
// Init/Start/Stop cannot run again. The next newGlobalId call returns maxId + 1.
func Init(timerList []Timer) {
	state := lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()
	if state != idle {
		panic(fmt.Sprintf("timer: Init in state %s", state))
	}

	var maxId int64
	for _, item := range timerList {
		if item.Id > maxId {
			maxId = item.Id
		}
		completeTimerPtr.addTimer(item)
	}
	setGlobalId(maxId)
	currPeriod := newPeriod(uint64(time.Now().UnixMilli()), periodIntervalMs)
	initTimerUntilCurrPeriod(currPeriod)
	lifecycleMutex.SetDataAfterLock(inited)
}

// Start launches the scheduler goroutine.
//
// It panics unless the lifecycle is inited (after Init, before Start).
func Start() {
	state := lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()
	if state != inited {
		panic(fmt.Sprintf("timer: Start in state %s", state))
	}
	lifecycleMutex.SetDataAfterLock(running)
	go schedule()
}

// Stop tells the scheduler to exit and waits until it does.
// The result is every timer still stored in completeTimerPtr for persistence.
// HandleTimer no-ops once Stop has moved the lifecycle to stopped, so in-flight
// channel copies cannot remove those entries afterward.
//
// It panics unless the lifecycle is running. After the scheduler exits, Stop clears
// currPeriodTimerList, closes DueTimerChannel, and moves the lifecycle to stopped.
// Init/Start/Stop cannot run again.
func Stop() []Timer {
	state := lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()
	if state != running {
		panic(fmt.Sprintf("timer: Stop in state %s", state))
	}

	closingChannel <- struct{}{}
	<-closedChannel
	clearCurrPeriodTimerList()
	close(DueTimerChannel)
	lifecycleMutex.SetDataAfterLock(stopped)
	return completeTimerPtr.getTimerList()
}

// schedule sends due timers to DueTimerChannel and advances the current period.
// It exits when Stop signals.
func schedule() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-closingChannel:
			closedChannel <- struct{}{}
			return
		case <-ticker.C:
			if !scheduleTick(time.Now().UnixMilli()) {
				closedChannel <- struct{}{}
				return
			}
		}
	}
}

// scheduleTick sends due timers and advances the current period when needed.
//
// When currPeriod equals globalCurrPeriod, it only peels and sends.
// When the period advances, it sends the old due list, rebuilds the current list
// from the map, and leaves new-period delivery to the next tick.
// CreateTimer only inserts timers for globalCurrPeriod, so a timer created in a
// new minute before the advance is loaded by the rebuild and sent on the next tick.
// It returns false when Stop interrupts a send; the caller signals closedChannel.
func scheduleTick(nowMs int64) bool {
	currPeriod := newPeriod(uint64(nowMs), periodIntervalMs)
	prevPeriod := getGlobalCurrPeriod()
	diffPeriod := currPeriod.GetPeriodDifference(prevPeriod)

	// Peel dues under currPeriodTimerMutex; send without holding it so CreateTimer
	// is not blocked on a full DueTimerChannel.
	if !sendDueTimers(takeDueTimers(nowMs)) {
		return false
	}
	if diffPeriod == 0 {
		return true
	}

	// Forward wall-clock advance rebuilds the due list. A step backward skips rebuild
	// on purpose (see package doc); globalCurrPeriod is left unchanged.
	if currPeriod.IsLaterThan(prevPeriod) {
		switch {
		case diffPeriod == 1:
			initTimerOfCurrPeriod(currPeriod)
		case diffPeriod > 1:
			initTimerBetweenPeriod(prevPeriod, currPeriod)
		}
	}

	return true
}

// sendDueTimers pushes due timers onto DueTimerChannel.
// A send waits in a select with closingChannel, so Stop can interrupt a full buffer.
// It returns false when Stop wins that select. The caller signals closedChannel.
func sendDueTimers(dueList []Timer) bool {
	for _, item := range dueList {
		select {
		case DueTimerChannel <- item:
		case <-closingChannel:
			return false
		}
	}
	return true
}

// GetTimer returns the timer matching briefTimer, or none when it is absent.
func GetTimer(briefTimer BriefTimer) Option[Timer] {
	return completeTimerPtr.getTimer(briefTimer)
}

// GetTimerCount returns how many timers are still stored in the package collection.
func GetTimerCount() int {
	return completeTimerPtr.getTimerCount()
}

// GetTimerList returns every timer still stored in the package collection.
// External systems can persist the list and pass it to Init; the id is derived
// as the maximum Id in that list.
func GetTimerList() []Timer {
	return completeTimerPtr.getTimerList()
}

// GetTimerListByType returns every stored timer whose TimerType equals timerType.
func GetTimerListByType(timerType string) []Timer {
	return completeTimerPtr.getTimerListByType(timerType)
}

// CreateTimer builds a timer that ends delayMs after now, stores it, and returns it.
// When the lifecycle is not running it returns a zero Timer without storing.
//
// It panics when delayMs is negative or when beginMs+delayMs overflows int64.
// When the timer ends in globalCurrPeriod, it is also inserted into the sorted
// current-period due list.
func CreateTimer(timerType string, delayMs int64, intData1, intData2, intData3 int64, stringData string) Timer {
	var timer Timer
	lifecycleMutex.ReadData(func(state lifecycleState) {
		if state != running {
			return
		}
		if delayMs < 0 {
			panic(fmt.Sprintf("timer: delayMs %d is negative", delayMs))
		}

		beginMs := time.Now().UnixMilli()
		if delayMs > math.MaxInt64-beginMs {
			panic(fmt.Sprintf("timer: delayMs %d overflows end time", delayMs))
		}
		endMs := beginMs + delayMs
		timer = newTimer(newGlobalId(), timerType, beginMs, endMs, intData1, intData2, intData3, stringData)
		storeTimerForSchedule(timer)
	})
	return timer
}

// CancelTimer removes a valid timer from the package collection. An invalid timer
// is ignored. When the lifecycle is not running it is a no-op.
// The matching brief, if any, stays on currPeriodTimerList until the next due peel,
// which drops it without sending (lazy peel).
func CancelTimer(timer Timer) {
	lifecycleMutex.ReadData(func(state lifecycleState) {
		if state != running {
			return
		}
		if !timer.IsValid() {
			return
		}
		completeTimerPtr.deleteTimer(timer)
	})
}

// HandleTimer runs the handler registered for timer.TimerType when timer is still
// present in completeTimerPtr, then removes it. A missing timer (for example after
// CancelTimer) is a no-op. When the timer is present it is removed even if no
// handler is registered. When the lifecycle is not running it returns nil.
//
// It returns an error when timer.TimerType has no handler.
func HandleTimer(timer Timer) error {
	var err error
	lifecycleMutex.ReadData(func(state lifecycleState) {
		if state != running {
			return
		}
		err = handleTimer(timer)
	})
	return err
}

func handleTimer(timer Timer) error {
	if completeTimerPtr.getTimer(timer.BriefTimer).HasNone() {
		return nil
	}

	defer func() {
		completeTimerPtr.deleteTimer(timer)
	}()

	handlerOption := getHandler(timer.TimerType)
	if handlerOption.HasNone() {
		return fmt.Errorf("timer: no handler for timer type %s", timer.TimerType)
	}

	handlerOption.Data()(timer)
	return nil
}
