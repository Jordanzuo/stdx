package timer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkPeriodNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newPeriod(uint64(i)*1000, periodIntervalMs)
	}
}

func BenchmarkPeriodGetPeriodDifference(b *testing.B) {
	a := newPeriod(0, periodIntervalMs)
	c := newPeriod(10*periodIntervalMs, periodIntervalMs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.GetPeriodDifference(a)
	}
}

func BenchmarkBriefTimerGetElapsedMs(b *testing.B) {
	now := time.Now().UnixMilli()
	timer := NewBriefTimer(1, now-500, now+500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = timer.GetElapsedMs()
	}
}

func BenchmarkCompleteTimerAddTimer(b *testing.B) {
	complete := newCompleteTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		complete.addTimer(newTimer(int64(i+1), "bench", 1000, 2000, 0, 0, 0, ""))
	}
}

func BenchmarkCompleteTimerAddTimerParallel(b *testing.B) {
	complete := newCompleteTimer()
	var id atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			next := id.Add(1)
			complete.addTimer(newTimer(next, "bench", 1000, 2000, 0, 0, 0, ""))
		}
	})
}

func BenchmarkCompleteTimerGetTimer(b *testing.B) {
	complete := newCompleteTimer()
	timer := newTimer(1, "bench", 1000, 2000, 0, 0, 0, "")
	complete.addTimer(timer)
	brief := timer.BriefTimer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = complete.getTimer(brief)
	}
}

func BenchmarkCompleteTimerGetTimerList(b *testing.B) {
	complete := newCompleteTimer()
	for i := int64(1); i <= 1000; i++ {
		end := int64((i%10)+1) * periodIntervalMs / 2
		complete.addTimer(newTimer(i, "bench", 0, end, 0, 0, 0, ""))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = complete.getTimerList()
	}
}

func BenchmarkAddTimerToCurrPeriodList(b *testing.B) {
	resetCurrPeriodTimerList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%1024 == 0 {
			currPeriodTimerList = currPeriodTimerList[:0]
		}
		addTimerToCurrPeriodList(NewBriefTimer(int64(i+1), 0, int64(b.N-i)))
	}
}

func BenchmarkGetDueTimerList(b *testing.B) {
	completeTimerPtr = newCompleteTimer()
	resetCurrPeriodTimerList()
	now := time.Now().UnixMilli()
	globalCurrPeriod = newPeriod(uint64(now), periodIntervalMs)
	for i := int64(1); i <= 256; i++ {
		timer := newTimer(i, "bench", now-1000, now-int64(i), 0, 0, 0, "")
		completeTimerPtr.addTimer(timer)
		addTimerToCurrPeriodList(timer.BriefTimer)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := getDueTimerList(now)
		for _, item := range list {
			addTimerToCurrPeriodList(item.BriefTimer)
		}
	}
}

func BenchmarkScheduleTickSteady(b *testing.B) {
	drainDueTimerChannel()
	completeTimerPtr = newCompleteTimer()
	resetCurrPeriodTimerList()
	now := time.Now().UnixMilli()
	globalCurrPeriod = newPeriod(uint64(now), periodIntervalMs)
	for i := int64(1); i <= 64; i++ {
		timer := newTimer(i, "bench", now, now+int64(i)*10, 0, 0, 0, "")
		completeTimerPtr.addTimer(timer)
		addTimerToCurrPeriodList(timer.BriefTimer)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduleTick(now)
		drainDueTimerChannel()
		currPeriodTimerList = currPeriodTimerList[:0]
		for _, item := range completeTimerPtr.getTimerList() {
			addTimerToCurrPeriodList(item.BriefTimer)
		}
	}
}

func BenchmarkScheduleTickPeriodAdvance(b *testing.B) {
	drainDueTimerChannel()
	completeTimerPtr = newCompleteTimer()
	now := time.Now().UnixMilli()
	curr := newPeriod(uint64(now), periodIntervalMs)
	prev := curr.GetPreviousPeriod()
	for i := int64(1); i <= 64; i++ {
		end := int64(curr.Interval.Lower) + i
		timer := newTimer(i, "bench", end-1, end, 0, 0, 0, "")
		completeTimerPtr.addTimer(timer)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalCurrPeriod = prev
		currPeriodTimerList = currPeriodTimerList[:0]
		scheduleTick(now)
		drainDueTimerChannel()
	}
}

func BenchmarkCreateTimer(b *testing.B) {
	resetSchedulerForTest()
	completeTimerPtr = newCompleteTimer()
	Init(nil)
	Start()
	defer resetSchedulerForTest()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateTimer("bench", periodIntervalMs, 0, 0, 0, "")
	}
}

var registerBenchHandleOnce sync.Once

func BenchmarkHandleTimer(b *testing.B) {
	registerBenchHandleOnce.Do(func() {
		RegisterHandler("bench-handle", func(Timer) {})
	})
	completeTimerPtr = newCompleteTimer()
	lifecycleMutex.SetData(running)
	defer lifecycleMutex.SetData(idle)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := newTimer(int64(i+1), "bench-handle", 1000, 2000, 1, 2, 3, "x")
		completeTimerPtr.addTimer(timer)
		if err := HandleTimer(timer); err != nil {
			b.Fatal(err)
		}
	}
}
