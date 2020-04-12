package namedsql

// ArrangeAndExpand performs Arrange followed by Expand, but parses the query
// only once.
func ArrangeAndExpand(query string, bindings map[string]interface{}, positionals ...interface{}) (string, []interface{}, error) {
	tokens := Lex(query)
	tokens, positionals, err := arrange(tokens, bindings, positionals...)
	if err != nil {
		return "", nil, err
	}

	tokens, positionals, err = expand(tokens, positionals...)
	if err != nil {
		return "", nil, err
	}

	return Render(tokens), positionals, nil
}

// MustArrangeAndExpand forwards to ArrangeAndExpand, except that its return
// values omit the trailing error and instead MustArrangeAndExpand panics on
// error.
func MustArrangeAndExpand(query string, bindings map[string]interface{}, positionals ...interface{}) (string, []interface{}) {
	query, positionals, err := ArrangeAndExpand(query, bindings, positionals...)
	if err != nil {
		panic(err)
	}

	return query, positionals
}
