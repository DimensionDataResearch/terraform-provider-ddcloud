package retry

import (
	"fmt"
	"log"
)

// Context represents contextual information about the current iteration of a retryable operation.
type Context interface {
	// Write a formatted message to the log.
	Log(format string, v ...interface{})

	// Retry the operation once the current iteration completes.
	Retry()

	// Mark the current iteration as failed due to the specified error.
	Fail(err error)
}

// Create a new runnerContext.
func newRunnerContext(operationDescription string) *runnerContext {
	return &runnerContext{
		OperationDescription: operationDescription,
		IterationCount:       0,
		ShouldRetry:          false,
		Error:                nil,
	}
}

type runnerContext struct {
	OperationDescription string
	IterationCount       int
	ShouldRetry          bool
	Error                error
}

var _ Context = &runnerContext{}

// Write a formatted message to the log.
func (context *runnerContext) Log(format string, formatArgs ...interface{}) {
	log.Printf(format, formatArgs...)
}

// Retry the operation once the current iteration completes.
func (context *runnerContext) Retry() {
	context.ShouldRetry = true
}

// Mark the current iteration as failed due to the specified error.
func (context *runnerContext) Fail(err error) {
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
func (context *runnerContext) NextIteration() {
	context.ShouldRetry = false
	context.Error = nil
	context.IterationCount++
}
