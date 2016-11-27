package retry

import (
	"fmt"
	"time"
)

// IsTimeoutError determines whether the specified error represents an operation timeout.
func IsTimeoutError(err error) bool {
	_, ok := err.(*OperationTimeoutError)

	return ok
}

// OperationTimeoutError is raised when the timeout for an operation is exceeded.
type OperationTimeoutError struct {
	// The operation description.
	OperationDescription string

	// The operation timeout period.
	Timeout time.Duration

	// The number of attempts that were made to perform the operation.
	Attempts int
}

// Error creates a string representation of the OperationTimeoutError.
func (timeoutError *OperationTimeoutError) Error() string {
	return fmt.Sprintf("%s - operation timed out after %d seconds (%d attempts)",
		timeoutError.OperationDescription,
		timeoutError.Timeout/time.Second,
		timeoutError.Attempts,
	)
}

var _ error = &OperationTimeoutError{}
