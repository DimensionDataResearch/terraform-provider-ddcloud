package assert

import (
	"reflect"
	"testing"
)

// Helper performs assertions for a specific test.
type Helper interface {
	// GetTest retrieves the test for which assertion facilities are being provided.
	GetTest() *testing.T

	// IsTrue asserts that the specified condition is true.
	IsTrue(description string, condition bool)

	// IsFalse asserts that the specified condition is true.
	IsFalse(description string, condition bool)

	// IsNil asserts that the specified value is nil.
	IsNil(description string, actual interface{})

	// IsNil asserts that the specified value is not nil.
	NotNil(description string, actual interface{})

	// Equals asserts that the specified values are equal.
	//
	// Uses reflect.DeepEquals.
	Equals(description string, expected interface{}, actual interface{})

	// EqualsString asserts that the specified string values are equal.
	EqualsString(description string, expected string, actual string)

	// EqualsInt asserts that the specified int values are equal.
	EqualsInt(description string, expected int, actual int)
}

type assertHelper struct {
	test *testing.T
}

// ForTest creates a new assertion Helper for the specified test.
func ForTest(test *testing.T) Helper {
	return assertHelper{test}
}

// GetTest retrieves the test for which assertion facilities are being provided.
func (assert assertHelper) GetTest() *testing.T {
	return assert.test
}

// IsTrue asserts that the specified condition is true.
func (assert assertHelper) IsTrue(description string, condition bool) {
	if !condition {
		assert.test.Fatalf("Expression was false: %s", description)
	}
}

// IsFalse asserts that the specified condition is true.
func (assert assertHelper) IsFalse(description string, condition bool) {
	if condition {
		assert.test.Fatalf("Expression was true: %s", description)
	}
}

// IsNil asserts that the specified value is nil.
func (assert assertHelper) IsNil(description string, actual interface{}) {
	if !reflect.ValueOf(actual).IsNil() {
		assert.test.Fatalf("%s was not nil.", description)
	}
}

// IsNil asserts that the specified value is not nil.
func (assert assertHelper) NotNil(description string, actual interface{}) {
	if reflect.ValueOf(actual).IsNil() {
		assert.test.Fatalf("%s was nil.", description)
	}
}

// Equals asserts that the specified values are equal.
//
// Uses reflect.DeepEqual.
func (assert assertHelper) Equals(description string, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		assert.test.Fatalf("%s was '%#v' (expected '%#v').", description, actual, expected)
	}
}

// EqualsString asserts that the specified string values are equal.
func (assert assertHelper) EqualsString(description string, expected string, actual string) {
	if actual != expected {
		assert.test.Fatalf("%s was '%s' (expected '%s').", description, actual, expected)
	}
}

// EqualsInt asserts that the specified int values are equal.
func (assert assertHelper) EqualsInt(description string, expected int, actual int) {
	if actual != expected {
		assert.test.Fatalf("%s was %d (expected %d).", description, actual, expected)
	}
}
