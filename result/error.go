package result

import "errors"

var (
	errNoData         = errors.New("result: no data")
	errNilError       = errors.New("result: nil err")
	errNotInitialized = errors.New("result: not initialized")
	errInvalidState   = errors.New("result: invalid state")
)

func ensureErr(err error) {
	if err == nil {
		panic(errNilError)
	}
}
