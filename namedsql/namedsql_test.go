package namedsql

// Here are unit test specific utilities used by two or more unit test files.

import "fmt"

// sliceCheck exists so we don't have to wonder which of "actual" and
// "expected" comes first.  Just use a sliceCheck instance instead.
type sliceCheck struct {
	actual   []interface{}
	expected []interface{}
}

// sliceDisagreement returns a diagnostic message if the actual and expected
// values in check are not equal, and returns an empty string if they are equal.
func sliceDisagreement(check sliceCheck) string {
	actual, expected := check.actual, check.expected

	if len(actual) != len(expected) {
		return fmt.Sprintf(
			"actual and expected slices have different numbers of elements.  "+
				"Expected slice has %d, but the actual slice has %d.\n"+
				"Expected: %v\n"+
				"Actual: %v",
			len(expected), len(actual), expected, actual)
	}

	for i, actualValue := range actual {
		expectedValue := expected[i]
		if actualValue == expectedValue {
			continue
		}

		return fmt.Sprintf(
			"actual and expected slices differ in at least one element.  At "+
				"the first position where they differ (index %d):\n"+
				"Expected element: %v\n"+
				"Actual element: %v",
			i, expectedValue, actualValue)
	}

	return ""
}
