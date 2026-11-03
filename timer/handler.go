package timer

import (
	"fmt"

	. "github.com/Jordanzuo/stdx/mutex"
	. "github.com/Jordanzuo/stdx/option"
)

// handlerMapMutex stores timer type to callback under a read/write lock.
var handlerMapMutex = NewRWMutex(make(map[string]func(Timer), 64))

// RegisterHandler binds handler to timerType.
//
// It panics when timerType already has a handler.
func RegisterHandler(timerType string, handler func(Timer)) {
	handlerMapMutex.WriteData(func(handlerMap map[string]func(Timer)) {
		if _, exists := handlerMap[timerType]; exists {
			panic(fmt.Sprintf("timer: %s has been registered", timerType))
		}
		handlerMap[timerType] = handler
	})
}

// getHandler returns the handler registered for timerType, or none when absent.
func getHandler(timerType string) Option[func(Timer)] {
	var handlerOption Option[func(Timer)]
	handlerMapMutex.ReadData(func(handlerMap map[string]func(Timer)) {
		handler, exists := handlerMap[timerType]
		if !exists {
			handlerOption = NewNoneOption[func(Timer)]()
			return
		}
		handlerOption = NewDataOption(handler)
	})
	return handlerOption
}
