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
func (do *doWithRetry) GetRetryPeriod() time.Duration {
	log.Printf("Do.GetRetryPeriod - stateLock.Lock()")
	do.stateLock.Lock()
	defer log.Printf("Do.GetRetryPeriod - stateLock.Unlock()")
	defer do.stateLock.Unlock()

	return do.retryPeriod
}

// SetRetryPeriod configures the Do's retry period.
//
// This determines how long the Do will wait between retries operations.
func (do *doWithRetry) SetRetryPeriod(retryPeriod time.Duration) {
	log.Printf("Do.SetRetryPeriod - stateLock.Lock()")
	do.stateLock.Lock()
	defer log.Printf("Do.SetRetryPeriod - stateLock.Unlock()")
	defer do.stateLock.Unlock()

	do.retryPeriod = retryPeriod
}

// DoAction performs the specified action until it succeeds or times out.
//
// description is a short description of the function used for logging.
// timeout is the period of time before the process
// action is the action function to invoke
//
// Returns the error (if any) passed to Context.Fail or caused by the operation timing out.
func (do *doWithRetry) Action(description string, timeout time.Duration, action ActionFunc) error {
	log.Printf("Do.Action - stateLock.Lock()")

	// Capture current configuration
	do.stateLock.Lock()
	retryPeriod := do.retryPeriod
	do.stateLock.Unlock()

	log.Printf("Do.Action - stateLock.Unlock()")

	// Perform the initial attempt immediately.
	initialAttemptTicker := make(chan bool, 1)
	initialAttemptTicker <- true

	waitTimeout := time.NewTimer(timeout)
	defer waitTimeout.Stop()

	retryTicker := time.NewTicker(retryPeriod)
	defer retryTicker.Stop()

	log.Printf("%s - will attempt operation once every %d seconds until successful (timeout after %d seconds)...",
		description,
		retryPeriod/time.Second,
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

		case <-initialAttemptTicker:
			context.NextIteration()

			log.Printf("%s - performing initial attempt...", description)

			action(context)
			if context.Error != nil {
				log.Printf("%s - initial attempt failed: %s.",
					description,
					context.Error,
				)

				return context.Error
			}

			if context.ShouldRetry {
				log.Printf("%s - initial attempt marked for retry (will try again)...", description)

				continue
			}

			log.Printf("%s - operation sucessful on initial attempt.", description)

			return nil

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

			log.Printf("%s - operation sucessful after %d attempts.",
				description,
				context.IterationCount,
			)

			return nil
		}
	}
}
