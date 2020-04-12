package namedsql

import (
	"fmt"
	"testing"
)

// Note that tokensCheck and tokensDisagreement are nearly copy paste from
// sliceCheck and sliceDisagreement defined in namedsql_test.go.  The
// duplication allows me to avoid having to cast a lot in these tests.

// tokensCheck exists so we don't have to wonder which of "actual" and
// "expected" comes first.  Just use a tokensCheck instance instead.
type tokensCheck struct {
	actual   []Token
	expected []Token
}

// tokensDisagreement returns a diagnostic message if the actual and expected
// values in check are not equal, and returns an empty string if they are
// equal.
func tokensDisagreement(check tokensCheck) string {
	actual, expected := check.actual, check.expected

	if len(actual) != len(expected) {
		return fmt.Sprintf(
			"Actual and expected values have different numbers of tokens.  "+
				"Expected value has %d tokens, but the actual value has %d.\n"+
				"Expected tokens: %v\n"+
				"Actual tokens: %v",
			len(expected), len(actual), expected, actual)
	}

	for i, actualToken := range actual {
		expectedToken := expected[i]
		if actualToken == expectedToken {
			continue
		}

		return fmt.Sprintf(
			"Actual and expected values differ in at least one token.  At "+
				"the first position where they differ (index %d):\n"+
				"Expected token: %v\n"+
				"Actual token: %v",
			i, expectedToken, actualToken)
	}

	return ""
}

func TestLexerLexBreathing1(t *testing.T) {
	// Here's an arbitrary test that I used as I was writing Lex.
	query := " -- foo\n/*bar*/NONSENSE'baz'\"buzz\"`fizz`?@1$2:wakka%(hah)s"
	expected := []Token{
		{Text: " "},        // non-matching
		{Text: "-- foo\n"}, // line comment
		{Text: "/*bar*/"},  // block comment
		{Text: "NONSENSE"}, // unmatched
		{Text: "'baz'"},    // single-quoted string
		{Text: `"buzz"`},   // double-quoted string
		{Text: "`fizz`"},   // back-quoted string
		{Kind: "implicit", Text: "?", Inside: "?"},       // implicit positional parameter
		{Kind: "explicit", Text: "@1", Inside: "1"},      // explicit positional parameter
		{Kind: "explicit", Text: "$2", Inside: "2"},      // explicit positional parameter
		{Kind: "named", Text: ":wakka", Inside: "wakka"}, // named parameter
		{Kind: "python", Text: "%(hah)s", Inside: "hah"}} // python-style named parameter
	tokens := Lex(query)
	message := tokensDisagreement(tokensCheck{actual: tokens, expected: expected})
	if message != "" {
		t.Error(message)
	}
}

func TestLexerLexBreathing2(t *testing.T) {
	// Here's another arbitrary test that I used as I was writing Lex.
	query := `-- Here's a more realistic example.
select foo, bar
from bazz
  inner join hah on bazz.id = hah.id
where foo = @some_damned_thing
  and bar in @more_things;`
	expected := []Token{
		{Text: "-- Here's a more realistic example.\n"}, // line comment
		{Text: "select foo, bar\nfrom bazz\n  inner join hah " +
			"on bazz.id = hah.id\nwhere foo = "}, // non-matching
		{Kind: "named", Text: "@some_damned_thing", Inside: "some_damned_thing"}, // named parameter
		{Text: "\n  and bar in "},                                    // non-matching
		{Kind: "named", Text: "@more_things", Inside: "more_things"}, // named parameter
		{Text: ";"}} // non-matching
	tokens := Lex(query)
	message := tokensDisagreement(tokensCheck{actual: tokens, expected: expected})
	if message != "" {
		t.Error(message)
	}
}

func TestLexerLexEmpty(t *testing.T) {
	// Empty query => no tokens
	query := ""
	expected := []Token{}
	tokens := Lex(query)
	message := tokensDisagreement(tokensCheck{actual: tokens, expected: expected})
	if message != "" {
		t.Error(message)
	}
}

func TestLexerLexBoring(t *testing.T) {
	// Nothing matches => one token
	query := "There's no parameter binding, strings, comments, or anything."
	expected := []Token{Token{Text: query}}
	tokens := Lex(query)
	message := tokensDisagreement(tokensCheck{actual: tokens, expected: expected})
	if message != "" {
		t.Error(message)
	}
}
