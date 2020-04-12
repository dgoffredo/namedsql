package namedsql

import (
	"fmt"
	"strconv"
)

// Arrange replaces all named parameters and explicit positional parameters in
// query with implicit positional parameters.  It returns the altered query and
// a slice of bindings corresponding to the parameters.  Named parameter
// bindings are looked up by name in the specified map, and positional bindings
// are looked up by index from among the optionally specified trailing
// arguments.
func Arrange(query string, bindings map[string]interface{}, positionals ...interface{}) (string, []interface{}, error) {
	tokens, positionals, err := arrange(Lex(query), bindings, positionals...)
	return Render(tokens), positionals, err
}

func arrange(tokens []Token, bindings map[string]interface{}, positionals ...interface{}) ([]Token, []interface{}, error) {
	outputTokens := make([]Token, 0, len(tokens))
	outputBindings := []interface{}{}
	nextPositionalIndex := 0

	appendParameter := func(binding interface{}) {
		// When we encounter a parameter in the input, we'll output a token and
		// a binding.
		outputTokens = append(outputTokens, Token{Kind: "implicit", Text: "?"})
		outputBindings = append(outputBindings, binding)
	}

	for _, token := range tokens {
		if token.Kind == "named" || token.Kind == "python" {
			// It's a named parameter.  Replace it with an implicit positional
			// parameter, and append the appropriate binding from `bindings`.
			name := token.Inside
			binding, ok := bindings[name]
			if !ok {
				whine := fmt.Errorf(
					"named parameter %q does not have a corresponding binding",
					token.Text)
				return nil, nil, whine
			}
			appendParameter(binding)
		} else if token.Kind == "explicit" {
			// It's an explicit positional parameter.  Replace it with an
			// implicit positional parameter, and append the appropriate
			// binding from `positionals`.
			i, err := strconv.Atoi(token.Inside)
			if err != nil {
				panic("I thought I understood regular expressions, but...")
			} else if i < 0 {
				panic("Now just wait a damned minute!")
			}

			if i == 0 {
				whine := fmt.Errorf(
					"invalid explicit positional parameter %q.  Index is one-based",
					token.Text)
				return nil, nil, whine
			}

			i--
			if i >= len(positionals) {
				whine := fmt.Errorf(
					"explicit positional parameter %q does not have a corresponding positional binding",
					token.Text)
				return nil, nil, whine
			}
			appendParameter(positionals[i])
		} else if token.Kind == "implicit" {
			// It's an implicit positional parameter.  Make sure that we
			// haven't run out of positional bindings, and then append the
			// appropriate parameter.
			if nextPositionalIndex == len(positionals) {
				whine := fmt.Errorf(
					"implicit positional parameter %q does not have a corresponding positional binding",
					token.Text)
				return nil, nil, whine
			}
			appendParameter(positionals[nextPositionalIndex])
			nextPositionalIndex++
		} else {
			// non-parameter tokens just get forwarded to the output
			outputTokens = append(outputTokens, token)
		}
	}

	return outputTokens, outputBindings, nil
}

// MustArrange forwards to Arrange, except that its return values omit the
// trailing error and instead MustArrange panics on error.
func MustArrange(query string, bindings map[string]interface{}, positionals ...interface{}) (string, []interface{}) {
	outputQuery, outputBindings, err := Arrange(query, bindings, positionals...)
	if err != nil {
		panic(err)
	}

	return outputQuery, outputBindings
}
