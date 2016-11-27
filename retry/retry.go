package retry

import "time"

// Default is the default Runner for retries.
var Default = NewRunner(30 * time.Second)

// DoAction performs the specified action until it succeeds or times out.
//
// description is a short description of the function used for logging.
// timeout is the period of time before the process
// action is the action function to invoke
//
// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
func DoAction(description string, timeout time.Duration, action ActionFunc) error {
	return Default.DoAction(description, timeout, action)
}
