package retry

import (
	"log"
	"sync"
	"time"
)

// Do is used to execute retriable operations.
type Do interface {
	// GetRetryPeriod retrieves the Do's currently-configured retry period.
	//
	// This determines how often the Do will retry operations.
	GetRetryPeriod() time.Duration

	// SetRetryPeriod configures the Do's retry period.
	//
	// This determines how long the Do will wait between retries operations.
	SetRetryPeriod(retryPeriod time.Duration)

	// DoAction performs the specified action until it succeeds or times out.
	//
	// description is a short description of the function used for logging.
	// timeout is the period of time before the process
	// action is the action function to invoke
	//
	// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
	Action(description string, timeout time.Duration, action ActionFunc) error
}

// NewDo creates a new Do.
func NewDo(retryPeriod time.Duration) Do {
	return &doWithRetry{
		stateLock:   &sync.Mutex{},
		retryPeriod: retryPeriod,
	}
}

type doWithRetry struct {
	stateLock   *sync.Mutex
	retryPeriod time.Duration
}

var _ Do = &doWithRetry{}

// GetRetryPeriod retrieves the Do's currently-configured retry period.
//
// This determines how often the Do will retry operations.
func (Do *doWithRetry) GetRetryPeriod() time.Duration {
	Do.stateLock.Lock()
	defer Do.stateLock.Unlock()

	return Do.retryPeriod
}

// SetRetryPeriod configures the Do's retry period.
//
// This determines how long the Do will wait between retries operations.
func (Do *doWithRetry) SetRetryPeriod(retryPeriod time.Duration) {
	Do.stateLock.Lock()
	defer Do.stateLock.Unlock()

	Do.retryPeriod = retryPeriod
}

// DoAction performs the specified action until it succeeds or times out.
//
// description is a short description of the function used for logging.
// timeout is the period of time before the process
// action is the action function to invoke
//
// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
func (Do *doWithRetry) Action(description string, timeout time.Duration, action ActionFunc) error {
	// Capture current configuration
	Do.stateLock.Lock()
	retryPeriod := Do.retryPeriod
	Do.stateLock.Unlock()

	waitTimeout := time.NewTimer(timeout)
	defer waitTimeout.Stop()

	retryTicker := time.NewTicker(retryPeriod)
	defer retryTicker.Stop()

	log.Printf("%s - will attempt operation once every %d seconds until successful (timeout after %d seconds)...",
		description,
		Do.retryPeriod/time.Second,
		timeout/time.Second,
	)

	context := newDoContext(description)
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
