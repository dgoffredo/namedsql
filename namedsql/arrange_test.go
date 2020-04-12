package namedsql

import "testing"

func TestArrangeBreathing(t *testing.T) {
	id := 1337
	color := "purple"
	limit := 10
	query, bindings, err := Arrange(
		"select * from llamas where id=? or friend=:1 and color=:color limit :limit;",
		map[string]interface{}{"color": color, "limit": limit},
		id)

	if err != nil {
		t.Errorf("error from Arrange: %v", err)
	}

	expectedQuery := "select * from llamas where id=? or friend=? and color=? limit ?;"
	if query != expectedQuery {
		t.Errorf("query not as expected.\nexpected: %q\nactual: %q", expectedQuery, query)
	}

	expectedBindings := []interface{}{id, id, color, limit}
	message := sliceDisagreement(sliceCheck{actual: bindings, expected: expectedBindings})
	if message != "" {
		t.Error(message)
	}
}
