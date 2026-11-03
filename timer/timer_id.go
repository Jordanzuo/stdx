package timer

import "sync/atomic"

// globalId is the last issued timer id. newGlobalId returns one greater than this value.
var globalId atomic.Int64

// setGlobalId sets the last issued timer id.
// The next newGlobalId call returns id + 1.
func setGlobalId(id int64) {
	globalId.Store(id)
}

// newGlobalId returns the next timer id. Concurrent calls receive distinct ids.
func newGlobalId() int64 {
	return globalId.Add(1)
}
