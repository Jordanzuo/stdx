package timer

import (
	"fmt"
	"sync"
	"testing"
)

func resetHandlerMap() {
	handlerMapMutex.SetData(make(map[string]func(Timer), 64))
}

func withRunningLifecycle(t *testing.T) {
	t.Helper()
	lifecycleMutex.SetData(running)
	t.Cleanup(func() { lifecycleMutex.SetData(idle) })
}

func TestRegisterHandler(t *testing.T) {
	resetHandlerMap()

	RegisterHandler("testType", func(Timer) {})

	var registered bool
	handlerMapMutex.ReadData(func(handlerMap map[string]func(Timer)) {
		_, registered = handlerMap["testType"]
	})
	if !registered {
		t.Fatal("handler should be registered")
	}
}

func TestRegisterHandlerConcurrent(t *testing.T) {
	resetHandlerMap()

	const count = 100
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		timerType := fmt.Sprintf("type-%d", i)
		go func() {
			defer wg.Done()
			RegisterHandler(timerType, func(Timer) {})
		}()
	}
	wg.Wait()

	var got int
	handlerMapMutex.ReadData(func(handlerMap map[string]func(Timer)) {
		got = len(handlerMap)
	})
	if got != count {
		t.Fatalf("handler count = %d, want %d", got, count)
	}
}

func TestRegisterHandlerDuplicate(t *testing.T) {
	resetHandlerMap()
	RegisterHandler("testType", func(Timer) {})

	defer func() {
		if recover() == nil {
			t.Fatal("RegisterHandler should panic on a duplicate registration")
		}
	}()
	RegisterHandler("testType", func(Timer) {})
}

func TestHandleTimer(t *testing.T) {
	resetHandlerMap()
	withRunningLifecycle(t)

	handlerCalled := false
	var receivedTimer Timer
	RegisterHandler("testType", func(timer Timer) {
		handlerCalled = true
		receivedTimer = timer
	})

	completeTimerPtr = NewCompleteTimer()
	timer := newTimer(1, "testType", 1000, 2000, 0, 0, 0, "")
	completeTimerPtr.addTimer(timer)

	if err := HandleTimer(timer); err != nil {
		t.Fatalf("HandleTimer() error = %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should be called")
	}
	if receivedTimer.Id != 1 {
		t.Fatalf("received timer id = %d, want 1", receivedTimer.Id)
	}
	if completeTimerPtr.GetTimer(NewBriefTimer(1, 1000, 2000)).HasData() {
		t.Fatal("timer should be deleted after HandleTimer")
	}
}

func TestHandleTimerNoHandler(t *testing.T) {
	resetHandlerMap()
	withRunningLifecycle(t)
	completeTimerPtr = NewCompleteTimer()
	timer := newTimer(1, "nonexistentType", 1000, 2000, 0, 0, 0, "")
	completeTimerPtr.addTimer(timer)

	if err := HandleTimer(timer); err == nil {
		t.Fatal("HandleTimer() should return an error for an unregistered type")
	}
	if completeTimerPtr.GetTimer(NewBriefTimer(1, 1000, 2000)).HasData() {
		t.Fatal("timer should be deleted when no handler exists")
	}
}

func TestHandleTimerMissingTimer(t *testing.T) {
	resetHandlerMap()
	withRunningLifecycle(t)
	handlerCalled := false
	RegisterHandler("testType", func(Timer) { handlerCalled = true })

	completeTimerPtr = NewCompleteTimer()
	timer := newTimer(1, "testType", 1000, 2000, 0, 0, 0, "")

	if err := HandleTimer(timer); err != nil {
		t.Fatalf("HandleTimer() error = %v", err)
	}
	if handlerCalled {
		t.Fatal("handler should not run when timer is absent from completeTimerPtr")
	}
}

func TestHandleTimerNotRunning(t *testing.T) {
	resetHandlerMap()
	handlerCalled := false
	RegisterHandler("testType", func(Timer) { handlerCalled = true })
	lifecycleMutex.SetData(idle)
	defer lifecycleMutex.SetData(idle)

	completeTimerPtr = NewCompleteTimer()
	timer := newTimer(1, "testType", 1000, 2000, 0, 0, 0, "")
	completeTimerPtr.addTimer(timer)

	if err := HandleTimer(timer); err != nil {
		t.Fatalf("HandleTimer() error = %v", err)
	}
	if handlerCalled {
		t.Fatal("handler should not run when lifecycle is not running")
	}
	if completeTimerPtr.GetTimer(timer.BriefTimer).HasNone() {
		t.Fatal("timer should remain when HandleTimer is a no-op")
	}
}

func TestHandleTimerMultipleTypes(t *testing.T) {
	resetHandlerMap()
	withRunningLifecycle(t)

	handler1Called := false
	handler2Called := false
	RegisterHandler("type1", func(Timer) { handler1Called = true })
	RegisterHandler("type2", func(Timer) { handler2Called = true })

	completeTimerPtr = NewCompleteTimer()
	timer1 := newTimer(1, "type1", 1000, 2000, 0, 0, 0, "")
	timer2 := newTimer(2, "type2", 1000, 3000, 0, 0, 0, "")
	completeTimerPtr.addTimer(timer1)
	completeTimerPtr.addTimer(timer2)

	if err := HandleTimer(timer1); err != nil {
		t.Fatalf("HandleTimer(timer1) error = %v", err)
	}
	if err := HandleTimer(timer2); err != nil {
		t.Fatalf("HandleTimer(timer2) error = %v", err)
	}
	if !handler1Called || !handler2Called {
		t.Fatal("both handlers should be called")
	}
}

func TestHandleTimerWithData(t *testing.T) {
	resetHandlerMap()
	withRunningLifecycle(t)

	var receivedInt1, receivedInt2, receivedInt3 int64
	var receivedString string
	RegisterHandler("testType", func(timer Timer) {
		receivedInt1 = timer.IntData1
		receivedInt2 = timer.IntData2
		receivedInt3 = timer.IntData3
		receivedString = timer.StringData
	})

	completeTimerPtr = NewCompleteTimer()
	timer := newTimer(1, "testType", 1000, 2000, 123, 456, 789, "testData")
	completeTimerPtr.addTimer(timer)
	if err := HandleTimer(timer); err != nil {
		t.Fatalf("HandleTimer() error = %v", err)
	}

	if receivedInt1 != 123 || receivedInt2 != 456 || receivedInt3 != 789 || receivedString != "testData" {
		t.Fatalf("got %d %d %d %s", receivedInt1, receivedInt2, receivedInt3, receivedString)
	}
}
