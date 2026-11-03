package timer

// Timer extends BriefTimer with a type name, three integer payloads, and a string payload.
//
// Tags name the same snake_case field in order json, bson, db, redis.
// BriefTimer is embedded without its own tag so those libraries keep id,
// begin_ms, and end_ms on the same record.
type Timer struct {
	// BriefTimer holds the id and the begin and end times.
	BriefTimer
	// TimerType identifies the timer category.
	TimerType string `json:"timer_type" bson:"timer_type" db:"timer_type" redis:"timer_type"`
	// IntData1 is the first integer payload.
	IntData1 int64 `json:"int_data1" bson:"int_data1" db:"int_data1" redis:"int_data1"`
	// IntData2 is the second integer payload.
	IntData2 int64 `json:"int_data2" bson:"int_data2" db:"int_data2" redis:"int_data2"`
	// IntData3 is the third integer payload.
	IntData3 int64 `json:"int_data3" bson:"int_data3" db:"int_data3" redis:"int_data3"`
	// StringData is the string payload.
	StringData string `json:"string_data" bson:"string_data" db:"string_data" redis:"string_data"`
}

// newTimer returns a Timer with the given id, type, bounds, and payloads.
func newTimer(id int64, timerType string, beginMs, endMs, intData1, intData2, intData3 int64, stringData string) Timer {
	return Timer{
		BriefTimer: NewBriefTimer(id, beginMs, endMs),
		TimerType:  timerType,
		IntData1:   intData1,
		IntData2:   intData2,
		IntData3:   intData3,
		StringData: stringData,
	}
}
