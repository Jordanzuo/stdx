// Package timer provides a process-wide period-bucketed timer scheduler.
//
// Public types are Period, BriefTimer, and Timer. Period is an inclusive-exclusive
// interval [start, end) over uint64 milliseconds. BriefTimer holds id and begin/end
// times; Timer adds a type name and payloads. The package store and period buckets
// are unexported; mutate only through CreateTimer, CancelTimer, and HandleTimer.
//
// RegisterHandler binds a callback to a timer type. Init, Start, and Stop must run
// once in that order. Init loads timers from a list and sets the last issued id to
// the maximum Id in that list. GetTimer, GetTimerCount, GetTimerList, and
// GetTimerListByType inspect the store. GetTimerList and Stop return remaining
// timers for persistence; after Stop, HandleTimer is a no-op, so persist the Stop
// list (or GetTimerList) and reload with Init. CreateTimer, CancelTimer, and
// HandleTimer take effect only while running. CancelTimer removes the map entry
// immediately; any brief left on the due list is dropped on the next peel without
// sending (lazy peel). Start runs the scheduler until Stop, which clears the
// current-period list, closes DueTimerChannel, and leaves the lifecycle stopped.
//
// A wall-clock step backward does not rebuild the due list (globalCurrPeriod stays
// put) so in-flight channel deliveries are not duplicated. Timers created while the
// clock is behind that stuck period may not be scheduled until the next Init; this
// package does not try to correct NTP or manual clock steps.
package timer

import (
	"fmt"

	"github.com/Jordanzuo/stdx/math"
)

// Period represents a time interval with a start and end time in milliseconds.
// The interval is inclusive at the start and exclusive at the end [start, end).
type Period struct {
	// Interval stores the time range using uint64 millisecond timestamps.
	Interval math.Interval[uint64]
}

// newPeriod returns a Period containing nowMs and lasting length milliseconds.
// The start is aligned down to a multiple of length.
//
// It panics when length is zero or when the aligned end overflows uint64.
func newPeriod(nowMs, length uint64) Period {
	if length == 0 {
		panic("timer: period length is zero")
	}
	remainder := nowMs % length
	beginTime := nowMs - remainder
	endTime := beginTime + length
	return Period{
		Interval: newInclusiveExclusive(beginTime, endTime),
	}
}

// Contains reports whether nowMs falls within this Period.
func (this Period) Contains(nowMs uint64) bool {
	return this.Interval.Contains(nowMs)
}

// IsEarlierThan reports whether this Period starts before other.
func (this Period) IsEarlierThan(other Period) bool {
	return this.Interval.Lower < other.Interval.Lower
}

// IsLaterThan reports whether this Period ends after other.
func (this Period) IsLaterThan(other Period) bool {
	return this.Interval.Upper > other.Interval.Upper
}

// GetPreviousPeriod returns a Period of the same length immediately preceding this one.
//
// It panics when the previous start would wrap below zero.
func (this Period) GetPreviousPeriod() Period {
	length := this.Interval.Length()
	if this.Interval.Lower < length {
		panic("timer: previous period underflows")
	}
	beginTime := this.Interval.Lower - length
	endTime := beginTime + length
	return Period{
		Interval: newInclusiveExclusive(beginTime, endTime),
	}
}

// GetNextPeriod returns a Period of the same length immediately following this one.
//
// It panics when the next end overflows uint64.
func (this Period) GetNextPeriod() Period {
	length := this.Interval.Length()
	beginTime := this.Interval.Upper
	endTime := beginTime + length
	return Period{
		Interval: newInclusiveExclusive(beginTime, endTime),
	}
}

// GetPeriodDifference returns how many periods apart this Period and other are,
// based on their start times. The result is an absolute count.
func (this Period) GetPeriodDifference(other Period) int {
	length := this.Interval.Length()
	var delta uint64
	if this.Interval.Lower >= other.Interval.Lower {
		delta = this.Interval.Lower - other.Interval.Lower
	} else {
		delta = other.Interval.Lower - this.Interval.Lower
	}
	return int(delta / length)
}

// String formats the period as "[start-end)" using millisecond bounds.
func (this Period) String() string {
	return fmt.Sprintf("[%d-%d)", this.Interval.Lower, this.Interval.Upper)
}

func newInclusiveExclusive(begin, end uint64) math.Interval[uint64] {
	r := math.BuildInterval(begin, end, math.InclusiveExclusive)
	if r.HasError() {
		panic(r.Error())
	}
	return r.Data()
}
