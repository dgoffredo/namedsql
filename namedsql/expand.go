package namedsql

import (
	"fmt"
	"reflect"
)

// Expand transforms the specified SQL query and bindings in the following way:
// Any implicit positional parameters in the query that are bound to arrays or
// slices are replaced with lists of implicit positional parameters referring
// to the elements of the original bindings.  For example,
//
//     a, b, c := 1, 2, 3
//     outputQuery, outputBindings := Expand(
//             "select * from t where x in ? and y = ?",
//             []int{a, b, c},
//             "foo")
//
// leaves outputQuery with the value
//
//     select * from t where x in (?, ?, ?) and y = ?
//
// and outputBindings with the value
//
//     []interface{}{a, b, c, "foo"}
//
func Expand(query string, bindings ...interface{}) (string, []interface{}, error) {
	tokens, bindings, err := expand(Lex(query), bindings...)
	return Render(tokens), bindings, err
}

func expand(tokens []Token, bindings ...interface{}) ([]Token, []interface{}, error) {
	bindingIndex := 0 // how far along we are consuming `bindings`
	outputTokens := make([]Token, 0, len(tokens))
	outputBindings := make([]interface{}, 0, len(bindings))
	for _, token := range tokens {
		if token.Kind == "implicit" {
			// It's a parameter. If the value is a sequence (e.g. a slice),
			// replace the parameter "?" with a list of parameters
			// "(?, ?, ...)" that refer to the sequence's elements.
			if bindingIndex == len(bindings) {
				whine := fmt.Errorf(
					"implicit positional parameter %q does not have a corresponding positional binding",
					token.Text)
				return nil, nil, whine
			}
			binding := bindings[bindingIndex]
			elements, isSequence := unpackSequence(binding)
			if isSequence {
				outputTokens = append(outputTokens, parameterList(len(elements))...)
				outputBindings = append(outputBindings, elements...)
			} else {
				outputTokens = append(outputTokens, token)
				outputBindings = append(outputBindings, binding)
			}
			bindingIndex++
		} else if token.Kind == "explicit" {
			whine := fmt.Errorf(
				"explicit positional parameters are not allowed in Expand.  parameter: %q",
				token.Text)
			return nil, nil, whine
		} else {
			outputTokens = append(outputTokens, token)
		}
	}

	return outputTokens, outputBindings, nil
}

// MustExpand forwards to Expand, except that its return values omit the
// trailing error and instead MustExpand panics on error.
func MustExpand(query string, bindings ...interface{}) (string, []interface{}) {
	outputQuery, outputBindings, err := Expand(query, bindings...)
	if err != nil {
		panic(err)
	}

	return outputQuery, outputBindings
}

// unpackSequence uses reflection to inspect sequence.  If sequence is an array
// or a slice, then unpackSequence returns a slice whose elements refer to the
// elements of sequence, and returns true to indicate that sequence is indeed a
// sequence.  If sequence is not an array or a slice, then unpackSequence
// returns nil, and returns false to indicate that sequence is not a sequence.
func unpackSequence(sequence interface{}) ([]interface{}, bool) {
	kind := reflect.TypeOf(sequence).Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		return nil, false
	}

	// It's an array or a slice, so unpack it.
	value := reflect.ValueOf(sequence)
	count := value.Len()
	elements := make([]interface{}, count)
	for i := 0; i < count; i++ {
		elements[i] = value.Index(i).Interface()
	}

	return elements, true
}

// parameterList returns a slice of tokens that form a SQL list containing
// implicit positional parameters (question marks), separated by spaces.  count
// is the number of parameters in the list.  For example,
//
//     Render(parameterList(4))
//
// returns the string "(?, ?, ?, ?)".
func parameterList(count int) []Token {
	tokens := []Token{{Text: "("}}

	if count != 0 {
		tokens = append(tokens, Token{Kind: "implicit", Text: "?"})
		for count--; count != 0; count-- {
			tokens = append(tokens,
				Token{Text: ", "},
				Token{Kind: "implicit", Text: "?"})
		}
	}

	return append(tokens, Token{Text: ")"})
}
