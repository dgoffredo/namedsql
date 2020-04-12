package namedsql

import "testing"

func TestExpandBreathing(t *testing.T) {
	a, b, c := 1, 2, 3
	query, bindings, err := Expand(
		"select * from t where x in ? and y = ?",
		[]int{a, b, c},
		"foo")

	if err != nil {
		t.Error(err)
	}

	expectedQuery := "select * from t where x in (?, ?, ?) and y = ?"
	if query != expectedQuery {
		t.Errorf("query not as expected.\nexpected: %q\nactual: %q", expectedQuery, query)
	}

	expectedBindings := []interface{}{a, b, c, "foo"}
	message := sliceDisagreement(sliceCheck{actual: bindings, expected: expectedBindings})
	if message != "" {
		t.Error(message)
	}
}

func TestExpandEmptyList(t *testing.T) {
	empty := []interface{}{}
	query, bindings, err := Expand("select * from t where x in ?", empty)

	if err != nil {
		t.Error(err)
	}

	expectedQuery := "select * from t where x in ()"
	if query != expectedQuery {
		t.Errorf("query not as expected.\nexpected: %q\nactual: %q", expectedQuery, query)
	}

	message := sliceDisagreement(sliceCheck{actual: bindings, expected: empty})
	if message != "" {
		t.Error(message)
	}
}
