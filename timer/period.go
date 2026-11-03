// Package timer provides Period, BriefTimer, Timer, PeriodTimer, and CompleteTimer.
//
// A Period is an inclusive-exclusive interval [start, end) stored as
// math.Interval[uint64]. Create one with NewPeriod, which aligns the start to a
// multiple of the period length. NewEmptyPeriod returns a zero Period.
// Do not call other methods on it.
//
// Bounds are uint64 milliseconds because math.Interval only supports unsigned
// integers. Length must be greater than zero. GetPreviousPeriod requires the
// current start to be at least one length above zero so the previous start does
// not wrap.
//
// A BriefTimer identifies a timer by a positive id and stores its begin and end
// as Unix millisecond timestamps. Elapsed and remaining time use time.Now.
// Timer embeds BriefTimer and adds a type name, three integer payloads, and a string payload.
// PeriodTimer stores the timers whose end times fall in one Period.
// CompleteTimer stores those period collections across the whole timeline.
// RegisterHandler binds a callback to a timer type. HandleTimer runs it only when
// the lifecycle is running and the timer is still present in completeTimerPtr
// (so Cancel after enqueue is safe). Init, Start, and Stop must run once in that
// order. Init loads the package-level CompleteTimer from a timer list and sets the
// last issued id to the maximum Id in that list; GetCompleteTimer returns the
// collection for read-only inspection (add/delete are package-private).
// GetTimerList and Stop return the remaining timers for persistence. After Stop,
// HandleTimer is a no-op, so buffered DueTimerChannel copies cannot clear the store;
// persist the Stop list (or GetTimerList) and reload with Init.
// CreateTimer, CancelTimer, and HandleTimer take effect only while the lifecycle is
// running; otherwise they return without changing state. CancelTimer removes the
// map entry immediately; any brief left on the current-period due list is dropped
// on the next peel without sending (lazy peel).
// newGlobalId issues the next id. Start runs schedule until Stop, which clears the
// current-period list, closes DueTimerChannel, and leaves the lifecycle stopped.
// The current period keeps a sorted brief-timer list; CreateTimer inserts under
// the same mutex the scheduler uses to peel dues and rebuild.
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

// NewEmptyPeriod returns a zero Period.
// Do not call other methods on it.
func NewEmptyPeriod() Period {
	return Period{}
}

// NewPeriod returns a Period containing nowMs and lasting length milliseconds.
// The start is aligned down to a multiple of length.
//
// It panics when length is zero or when the aligned end overflows uint64.
func NewPeriod(nowMs, length uint64) Period {
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
