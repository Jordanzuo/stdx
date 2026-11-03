package timer

import "time"

// BriefTimer is a timer with an id and a begin and end time in Unix milliseconds.
// An id greater than 0 is valid. The timer is due when the current time reaches EndMs.
type BriefTimer struct {
	// Id is the timer identifier. It must be greater than 0 to be valid.
	Id int64 `json:"id" bson:"id" db:"id" redis:"id"`
	// BeginMs is the start time in milliseconds since the Unix epoch.
	BeginMs int64 `json:"begin_ms" bson:"begin_ms" db:"begin_ms" redis:"begin_ms"`
	// EndMs is the due time in milliseconds since the Unix epoch.
	EndMs int64 `json:"end_ms" bson:"end_ms" db:"end_ms" redis:"end_ms"`
}

// NewBriefTimer returns a BriefTimer with the given id and millisecond bounds.
func NewBriefTimer(id, beginMs, endMs int64) BriefTimer {
	return BriefTimer{
		Id:      id,
		BeginMs: beginMs,
		EndMs:   endMs,
	}
}

// IsValid reports whether the timer id is greater than 0.
func (this BriefTimer) IsValid() bool {
	return this.Id > 0
}

// IsDue reports whether nowMs is at or past EndMs.
func (this BriefTimer) IsDue(nowMs int64) bool {
	return nowMs >= this.EndMs
}

// getPeriod returns the minute-aligned Period that contains EndMs.
// EndMs must be non-negative; callers that accept external input validate first.
func (this BriefTimer) getPeriod() Period {
	return NewPeriod(uint64(this.EndMs), periodIntervalMs)
}

// GetDurationMs returns EndMs minus BeginMs.
func (this BriefTimer) GetDurationMs() int64 {
	return this.EndMs - this.BeginMs
}

// GetDurationSec returns the duration in whole seconds.
func (this BriefTimer) GetDurationSec() int64 {
	return this.GetDurationMs() / millisecondPerSecond
}

// GetElapsedMs returns how long the timer has been running, capped at its duration.
// An invalid timer, or one whose BeginMs is still in the future, returns 0.
// The current time is time.Now in Unix milliseconds.
func (this BriefTimer) GetElapsedMs() int64 {
	if !this.IsValid() {
		return 0
	}

	now := time.Now().UnixMilli()
	if now <= this.BeginMs {
		return 0
	}
	if now >= this.EndMs {
		now = this.EndMs
	}	
	return now - this.BeginMs
}

// GetElapsedSec returns the elapsed time in whole seconds.
func (this BriefTimer) GetElapsedSec() int64 {
	return this.GetElapsedMs() / millisecondPerSecond
}

// GetRemainMs returns the milliseconds left until EndMs.
// An invalid timer, or one whose EndMs is in the past, returns 0.
// The current time is time.Now in Unix milliseconds.
func (this BriefTimer) GetRemainMs() int64 {
	if !this.IsValid() {
		return 0
	}

	now := time.Now().UnixMilli()
	if now >= this.EndMs {
		return 0
	}

	return this.EndMs - now
}

// GetRemainSec returns the remaining time in whole seconds.
func (this BriefTimer) GetRemainSec() int64 {
	return this.GetRemainMs() / millisecondPerSecond
}

// orderByEndTime reports whether this timer ends before other.
func (this BriefTimer) orderByEndTime(other BriefTimer) bool {
	return this.EndMs < other.EndMs
}
