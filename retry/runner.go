package retry

import (
	"log"
	"sync"
	"time"
)

// Runner is used to execute retriable operations.
type Runner interface {
	// GetRetryPeriod retrieves the runner's currently-configured retry period.
	//
	// This determines how often the Runner will retry operations.
	GetRetryPeriod() time.Duration

	// SetRetryPeriod configures the runner's retry period.
	//
	// This determines how long the Runner will wait between retries operations.
	SetRetryPeriod(retryPeriod time.Duration)

	// DoAction performs the specified action until it succeeds or times out.
	//
	// description is a short description of the function used for logging.
	// timeout is the period of time before the process
	// action is the action function to invoke
	//
	// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
	DoAction(description string, timeout time.Duration, action ActionFunc) error
}

// NewRunner creates a new Runner.
func NewRunner(retryPeriod time.Duration) Runner {
	return &retryRunner{
		stateLock:   &sync.Mutex{},
		retryPeriod: retryPeriod,
	}
}

type retryRunner struct {
	stateLock   *sync.Mutex
	retryPeriod time.Duration
}

var _ Runner = &retryRunner{}

// GetRetryPeriod retrieves the runner's currently-configured retry period.
//
// This determines how often the Runner will retry operations.
func (runner *retryRunner) GetRetryPeriod() time.Duration {
	runner.stateLock.Lock()
	defer runner.stateLock.Unlock()

	return runner.retryPeriod
}

// SetRetryPeriod configures the runner's retry period.
//
// This determines how long the Runner will wait between retries operations.
func (runner *retryRunner) SetRetryPeriod(retryPeriod time.Duration) {
	runner.stateLock.Lock()
	defer runner.stateLock.Unlock()

	runner.retryPeriod = retryPeriod
}

// DoAction performs the specified action until it succeeds or times out.
//
// description is a short description of the function used for logging.
// timeout is the period of time before the process
// action is the action function to invoke
//
// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
func (runner *retryRunner) DoAction(description string, timeout time.Duration, action ActionFunc) error {
	// Capture current configuration
	runner.stateLock.Lock()
	retryPeriod := runner.retryPeriod
	runner.stateLock.Unlock()

	waitTimeout := time.NewTimer(timeout)
	defer waitTimeout.Stop()

	retryTicker := time.NewTicker(retryPeriod)
	defer retryTicker.Stop()

	log.Printf("%s - will attempt operation once every %d seconds until successful (timeout after %d seconds)...",
		description,
		runner.retryPeriod/time.Second,
		timeout/time.Second,
	)

	context := newRunnerContext(description)
	for {
		select {
		case <-waitTimeout.C:
			log.Printf("%s - operation timed out after %d seconds (%d attempts)",
				description,
				timeout/time.Second,
				context.IterationCount,
			)

			return &OperationTimeoutError{
				OperationDescription: description,
				Timeout:              timeout,
				Attempts:             context.IterationCount,
			}

		case <-retryTicker.C:
			context.NextIteration()

			log.Printf("%s - performing attempt %d...",
				description,
				context.IterationCount,
			)

			action(context)
			if context.Error != nil {
				log.Printf("%s - attempt %d failed: %s.",
					description,
					context.IterationCount,
					context.Error,
				)

				return context.Error
			}

			if context.ShouldRetry {
				log.Printf("%s - attempt %d marked for retry (will try again)...",
					description,
					context.IterationCount,
				)

				continue
			}

			log.Printf("%s - operation sucessful after %d attempt(s).",
				description,
				context.IterationCount,
			)

			return nil
		}
	}
}
