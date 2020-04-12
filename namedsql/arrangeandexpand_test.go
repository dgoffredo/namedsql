package namedsql

import "testing"

func TestArrangeAndExpandBreathing(t *testing.T) {
	gender, orientation, userID := 0, 1, "steve"
	query, bindings, err := ArrangeAndExpand(`
select value
from tags
where type in @types
  and userid = @userID`,
		map[string]interface{}{
			"types":  []int{gender, orientation},
			"userID": userID})

	if err != nil {
		t.Error(err)
	}

	expectedQuery := "\nselect value\nfrom tags\nwhere type in (?, ?)\n  and userid = ?"
	if query != expectedQuery {
		t.Errorf("query not as expected.\nexpected: %q\nactual: %q", expectedQuery, query)
	}

	expectedBindings := []interface{}{gender, orientation, userID}
	message := sliceDisagreement(sliceCheck{actual: bindings, expected: expectedBindings})
	if message != "" {
		t.Error(message)
	}
}
