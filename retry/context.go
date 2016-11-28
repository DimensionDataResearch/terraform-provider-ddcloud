package retry

import (
	"fmt"
	"log"
)

// Context represents contextual information about the current iteration of a retryable operation.
type Context interface {
	// Retry the operation once the current iteration completes.
	Retry()

	// Mark the current iteration as failed due to the specified error.
	Fail(err error)
}

// Create a new doContext.
func newDoContext(operationDescription string) *doContext {
	return &doContext{
		OperationDescription: operationDescription,
		IterationCount:       0,
		ShouldRetry:          false,
		Error:                nil,
	}
}

type doContext struct {
	OperationDescription string
	IterationCount       int
	ShouldRetry          bool
	Error                error
}

var _ Context = &doContext{}

// Retry the operation once the current iteration completes.
func (context *doContext) Retry() {
	context.ShouldRetry = true
}

// Mark the current iteration as failed due to the specified error.
func (context *doContext) Fail(err error) {
	context.Error = err

	if err != nil {
		iterationDescription := ""
		if context.IterationCount > 1 {
			iterationDescription = fmt.Sprintf(" (retry %d)",
				context.IterationCount,
			)
		}

		log.Printf("%s%s failed: %s",
			context.OperationDescription,
			iterationDescription,
			err,
		)
	}
}

// NextIteration resets the context for the next iteration.
func (context *doContext) NextIteration() {
	context.ShouldRetry = false
	context.Error = nil
	context.IterationCount++
}
