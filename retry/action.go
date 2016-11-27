package retry

// ActionFunc is a function representing a retryable operation that does not directly return any value.
//
// Feel free to publish values from the function to variables in the enclosing scope.
type ActionFunc func(context Context)
