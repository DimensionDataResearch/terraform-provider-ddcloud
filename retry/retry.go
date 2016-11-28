package retry

import "time"

// DefaultDo is the default executor for retrying operations.
var DefaultDo = NewDo(30 * time.Second)

// Action performs the specified action until it succeeds or times out.
//
// description is a short description of the function used for logging.
// timeout is the period of time before the process
// action is the action function to invoke
//
// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
func Action(description string, timeout time.Duration, action ActionFunc) error {
	return DefaultDo.Action(description, timeout, action)
}
