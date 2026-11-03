package timer

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	timerList := []Timer{
		newTimer(1, "type1", 1000, 2000, 0, 0, 0, ""),
		newTimer(2, "type2", 0, 70000, 0, 0, 0, ""),
	}
	Init(timerList)

	if got := newGlobalId(); got != 3 {
		t.Fatalf("next id = %d, want 3", got)
	}
	if GetCompleteTimer() != completeTimerPtr {
		t.Fatal("GetCompleteTimer should return the package collection")
	}
	if len(completeTimerPtr.GetTimerList()) != 2 {
		t.Fatalf("timer count = %d, want 2", len(completeTimerPtr.GetTimerList()))
	}
	if completeTimerPtr.GetTimer(NewBriefTimer(2, 0, 70000)).HasNone() {
		t.Fatal("timer 2 should be loaded")
	}
}

func TestGetTimerList(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	timerList := []Timer{
		newTimer(1, "type1", 1000, 2000, 0, 0, 0, "a"),
		newTimer(2, "type2", 0, 70000, 0, 0, 0, "b"),
	}
	Init(timerList)

	got := GetTimerList()
	if len(got) != 2 {
		t.Fatalf("timer count = %d, want 2", len(got))
	}
	found := map[int64]bool{}
	for _, item := range got {
		found[item.Id] = true
	}
	if !found[1] || !found[2] {
		t.Fatalf("timers = %+v, want ids 1 and 2", got)
	}
}

func TestCreateTimer(t *testing.T) {
	resetSchedulerForTest()
	Init(nil)
	Start()
	defer resetSchedulerForTest()
	timer := CreateTimer("type1", periodIntervalMs, 11, 22, 33, "payload")
	if timer.TimerType != "type1" || timer.IntData1 != 11 || timer.IntData2 != 22 || timer.IntData3 != 33 || timer.StringData != "payload" {
		t.Fatalf("got %+v", timer)
	}
	if completeTimerPtr.GetTimer(timer.BriefTimer).HasNone() {
		t.Fatal("created timer should be stored")
	}
}

func TestCreateTimerLaterPeriod(t *testing.T) {
	resetSchedulerForTest()
	Init(nil)
	Start()
	defer resetSchedulerForTest()

	timer := CreateTimer("type1", int64(periodIntervalMs), 0, 0, 0, "")
	if timer.getPeriod() == NewPeriod(uint64(timer.BeginMs), periodIntervalMs) {
		t.Fatal("delay of one period should end in a later period")
	}
	currPeriodTimerMutex.Lock()
	for _, item := range currPeriodTimerList {
		if item == timer.BriefTimer {
			currPeriodTimerMutex.Unlock()
			t.Fatalf("later period should not be on due list, got %#v", item)
		}
	}
	currPeriodTimerMutex.Unlock()
}

func TestCreateTimerNegativeDelay(t *testing.T) {
	resetSchedulerForTest()
	Init(nil)
	Start()
	defer resetSchedulerForTest()
	defer func() {
		if recover() == nil {
			t.Fatal("negative delay should panic")
		}
	}()
	CreateTimer("type1", -1, 0, 0, 0, "")
}

func TestCreateTimerOverflowDelay(t *testing.T) {
	resetSchedulerForTest()
	Init(nil)
	Start()
	defer resetSchedulerForTest()
	defer func() {
		if recover() == nil {
			t.Fatal("overflowing delay should panic")
		}
	}()
	CreateTimer("type1", math.MaxInt64, 0, 0, 0, "")
}

func TestCancelTimer(t *testing.T) {
	resetSchedulerForTest()
	timer := newTimer(1, "type1", 1000, 2000, 0, 0, 0, "")
	Init([]Timer{timer})
	Start()
	defer resetSchedulerForTest()

	CancelTimer(newTimer(0, "type1", 1000, 2000, 0, 0, 0, ""))
	if completeTimerPtr.GetTimer(timer.BriefTimer).HasNone() {
		t.Fatal("cancelling an invalid timer should leave the collection unchanged")
	}

	CancelTimer(timer)
	if completeTimerPtr.GetTimer(timer.BriefTimer).HasData() {
		t.Fatal("timer should be cancelled")
	}
}

func TestScheduleTickSendsBoundaryTimerOnce(t *testing.T) {
	completeTimerPtr = NewCompleteTimer()
	currPeriodTimerList = make([]BriefTimer, 0, 64)

	now := time.Now().UnixMilli()
	curr := NewPeriod(uint64(now), periodIntervalMs)
	globalCurrPeriod = curr.GetPreviousPeriod()

	// CreateTimer would not stage this on the temporary list while globalCurrPeriod
	// is still the previous minute; the advance loads it from the map, and the next
	// tick sends it once.
	timer := newTimer(901, "boundary", now, now, 0, 0, 0, "once")
	completeTimerPtr.addTimer(timer)

	scheduleTick(now)
	if _, ok := takeDueTimer(); ok {
		t.Fatal("advance tick should only rebuild, not send the new-period timer")
	}

	scheduleTick(now)
	got, ok := takeDueTimer()
	if !ok {
		t.Fatal("boundary timer was not sent")
	}
	if got.Id != timer.Id {
		t.Fatalf("sent id = %d, want %d", got.Id, timer.Id)
	}

	scheduleTick(now)
	if _, ok := takeDueTimer(); ok {
		t.Fatal("boundary timer was sent a second time")
	}
}

func TestScheduleTickSendsTimerLoadedByPeriodJump(t *testing.T) {
	completeTimerPtr = NewCompleteTimer()
	currPeriodTimerList = make([]BriefTimer, 0, 64)

	now := time.Now().UnixMilli()
	curr := NewPeriod(uint64(now), periodIntervalMs)
	globalCurrPeriod = curr.GetPreviousPeriod().GetPreviousPeriod()
	skipped := getGlobalCurrPeriod().GetNextPeriod()

	timer := newTimer(902, "skipped", int64(skipped.Interval.Lower), int64(skipped.Interval.Lower)+1, 0, 0, 0, "gap")
	completeTimerPtr.addTimer(timer)

	scheduleTick(now)
	if _, ok := takeDueTimer(); ok {
		t.Fatal("advance tick should only rebuild, not send the loaded timer")
	}

	scheduleTick(now)
	got, ok := takeDueTimer()
	if !ok {
		t.Fatal("timer in the skipped period was not sent")
	}
	if got.Id != timer.Id {
		t.Fatalf("sent id = %d, want %d", got.Id, timer.Id)
	}
	if _, ok := takeDueTimer(); ok {
		t.Fatal("skipped-period timer was sent a second time")
	}
}

func takeDueTimer() (Timer, bool) {
	select {
	case got, ok := <-DueTimerChannel:
		if !ok {
			return Timer{}, false
		}
		return got, true
	default:
		return Timer{}, false
	}
}

func drainDueTimerChannel() {
	for {
		select {
		case _, ok := <-DueTimerChannel:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func resetSchedulerForTest() {
	state := lifecycleMutex.Lock()
	lifecycleMutex.Unlock()
	switch state {
	case running:
		Stop()
		DueTimerChannel = make(chan Timer, dueTimerChannelCapacity)
	case stopped:
		DueTimerChannel = make(chan Timer, dueTimerChannelCapacity)
	case inited:
		drainDueTimerChannel()
	default:
		drainDueTimerChannel()
	}
	select {
	case <-closingChannel:
	default:
	}
	clearCurrPeriodTimerList()
	lifecycleMutex.SetData(idle)
}

func stopWithTimeout(t *testing.T) []Timer {
	t.Helper()
	done := make(chan []Timer, 1)
	go func() {
		done <- Stop()
	}()
	select {
	case got := <-done:
		return got
	case <-time.After(2 * time.Second):
		t.Fatal("Stop blocked")
	}
	return nil
}

func TestWorkflowInitStartCreateCancelStop(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	currPeriodTimerList = make([]BriefTimer, 0, 64)

	now := time.Now().UnixMilli()
	loaded := newTimer(10, "loaded", now-5000, now-1000, 7, 8, 9, "from-init")
	Init([]Timer{loaded})
	if got := globalId.Load(); got != 10 {
		t.Fatalf("global id = %d, want 10", got)
	}
	if completeTimerPtr.GetTimer(loaded.BriefTimer).HasNone() {
		t.Fatal("Init should store the loaded timer")
	}
	if got := getGlobalCurrPeriod(); !got.Contains(uint64(now)) {
		t.Fatalf("globalCurrPeriod = %s should contain %d", got, now)
	}

	Start()
	defer resetSchedulerForTest()

	due := CreateTimer("due", 0, 1, 2, 3, "fire-me")
	keep := CreateTimer("keep", periodIntervalMs, 4, 5, 6, "keep-me")
	cancel := CreateTimer("cancel", periodIntervalMs, 0, 0, 0, "cancel-me")
	if due.Id != 11 || keep.Id != 12 || cancel.Id != 13 {
		t.Fatalf("ids = %d %d %d, want 11 12 13", due.Id, keep.Id, cancel.Id)
	}

	CancelTimer(cancel)
	if completeTimerPtr.GetTimer(cancel.BriefTimer).HasData() {
		t.Fatal("CancelTimer should remove the cancelled timer")
	}
	if completeTimerPtr.GetTimer(keep.BriefTimer).HasNone() {
		t.Fatal("kept timer should remain after CancelTimer")
	}

	gotDue := false
	gotLoaded := false
	deadline := time.After(2 * time.Second)
	for !gotDue || !gotLoaded {
		select {
		case got := <-DueTimerChannel:
			switch got.Id {
			case due.Id:
				if got.StringData != "fire-me" {
					t.Fatalf("due payload = %+v", got)
				}
				gotDue = true
			case loaded.Id:
				if got.StringData != "from-init" {
					t.Fatalf("loaded payload = %+v", got)
				}
				gotLoaded = true
			default:
				t.Fatalf("unexpected due timer %+v", got)
			}
		case <-deadline:
			t.Fatalf("timed out waiting for due timers: due=%v loaded=%v", gotDue, gotLoaded)
		}
	}

	timerList := stopWithTimeout(t)

	foundKeep := false
	foundCancel := false
	for _, item := range timerList {
		if item.Id == keep.Id && item.StringData == "keep-me" {
			foundKeep = true
		}
		if item.Id == cancel.Id {
			foundCancel = true
		}
	}
	if !foundKeep {
		t.Fatalf("Stop timer list = %+v, want kept timer %d", timerList, keep.Id)
	}
	if foundCancel {
		t.Fatal("Stop timer list should not include the cancelled timer")
	}
}

func TestSendDueTimersStopsWhenChannelFull(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	currPeriodTimerList = make([]BriefTimer, 0, 64)

	now := time.Now().UnixMilli()
	globalCurrPeriod = NewPeriod(uint64(now), periodIntervalMs)
	for i := 0; i < dueTimerChannelCapacity; i++ {
		DueTimerChannel <- Timer{}
	}
	defer drainDueTimerChannel()

	timer := newTimer(903, "blocked", now, now, 0, 0, 0, "stop")
	completeTimerPtr.addTimer(timer)
	addTimerToCurrPeriodList(timer.BriefTimer)
	dueList := takeDueTimers(now)

	done := make(chan bool, 1)
	go func() {
		done <- sendDueTimers(dueList)
	}()

	time.Sleep(50 * time.Millisecond)
	closingChannel <- struct{}{}

	select {
	case ok := <-done:
		if ok {
			t.Fatal("sendDueTimers should return false when Stop interrupts")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sendDueTimers blocked on a full DueTimerChannel")
	}
}

func TestStartStopSchedule(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	currPeriodTimerList = make([]BriefTimer, 0, 64)
	Init(nil)
	Start()
	defer resetSchedulerForTest()

	kept := CreateTimer("later", periodIntervalMs, 0, 0, 0, "kept")
	created := CreateTimer("type1", 0, 1, 2, 3, "payload")

	select {
	case got := <-DueTimerChannel:
		if got.Id != created.Id || got.StringData != "payload" {
			t.Fatalf("got %+v, want id %d", got, created.Id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for a due timer")
	}

	timerList := stopWithTimeout(t)
	found := false
	for _, item := range timerList {
		if item.Id == kept.Id && item.StringData == "kept" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Stop timer list = %+v, want timer %d", timerList, kept.Id)
	}
}

func TestConcurrentCreateCancelHandleAndSchedule(t *testing.T) {
	resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()
	resetHandlerMap()
	RegisterHandler("stress", func(Timer) {})

	Init(nil)
	Start()
	defer resetSchedulerForTest()

	var consumerDone sync.WaitGroup
	consumerDone.Add(1)
	go func() {
		defer consumerDone.Done()
		for timer := range DueTimerChannel {
			_ = HandleTimer(timer)
		}
	}()

	const workers = 8
	const perWorker = 100
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				timer := CreateTimer("stress", int64(i%3)*10, 0, 0, 0, "")
				if !timer.IsValid() {
					continue
				}
				if i%4 == 0 {
					CancelTimer(timer)
				}
			}
		}()
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	_ = stopWithTimeout(t)
	consumerDone.Wait()
}

func TestLifecycleOrder(t *testing.T) {
	resetSchedulerForTest()
	defer resetSchedulerForTest()
	completeTimerPtr = NewCompleteTimer()

	assertPanics(t, func() { Start() }, "Start before Init")
	assertPanics(t, func() { Stop() }, "Stop before Start")
	if got := CreateTimer("x", 0, 0, 0, 0, ""); got.IsValid() {
		t.Fatal("CreateTimer before Init should return a zero timer")
	}
	CancelTimer(newTimer(1, "x", 0, 1, 0, 0, 0, ""))
	if err := HandleTimer(newTimer(1, "x", 0, 1, 0, 0, 0, "")); err != nil {
		t.Fatalf("HandleTimer before Init error = %v", err)
	}

	Init(nil)
	assertPanics(t, func() { Init(nil) }, "Init already called")
	assertPanics(t, func() { Stop() }, "Stop before Start")
	if got := CreateTimer("x", 0, 0, 0, 0, ""); got.IsValid() {
		t.Fatal("CreateTimer before Start should return a zero timer")
	}
	CancelTimer(newTimer(1, "x", 0, 1, 0, 0, 0, ""))
	if err := HandleTimer(newTimer(1, "x", 0, 1, 0, 0, 0, "")); err != nil {
		t.Fatalf("HandleTimer before Start error = %v", err)
	}

	Start()
	assertPanics(t, func() { Start() }, "Start already called")
	assertPanics(t, func() { Init(nil) }, "Init while running")
	_ = CreateTimer("ok", periodIntervalMs, 0, 0, 0, "")

	stopWithTimeout(t)
	assertPanics(t, func() { Init(nil) }, "Init after Stop")
	assertPanics(t, func() { Start() }, "Start after Stop")
	assertPanics(t, func() { Stop() }, "Stop after Stop")
	if got := CreateTimer("x", 0, 0, 0, 0, ""); got.IsValid() {
		t.Fatal("CreateTimer after Stop should return a zero timer")
	}
	CancelTimer(newTimer(1, "x", 0, 1, 0, 0, 0, ""))
	if err := HandleTimer(newTimer(1, "x", 0, 1, 0, 0, 0, "")); err != nil {
		t.Fatalf("HandleTimer after Stop error = %v", err)
	}
}

func assertPanics(t *testing.T, fn func(), name string) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s should panic", name)
		}
	}()
	fn()
}
