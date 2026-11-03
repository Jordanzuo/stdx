package timer

const (
	// millisecondPerSecond is the number of milliseconds in one second.
	millisecondPerSecond = 1000
	// periodIntervalMs is the schedule period length in milliseconds (one minute).
	periodIntervalMs = 60 * millisecondPerSecond
	// dueTimerChannelCapacity is the buffer size of DueTimerChannel.
	dueTimerChannelCapacity = 1024
)
