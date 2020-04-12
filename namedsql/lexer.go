package namedsql

import (
	"regexp"
	"strings"
	"sync"
)

const (
	// natural is a regular expression pattern that matches either zero, or
	// some digits not starting with zero
	natural = `0|[1-9][0-9]*`

	// identifier is a regular expression pattern that matches a letter or an
	// underscore, followed by letters, underscores, or digits
	identifier = `(?:\pL|_)(?:\pL|\p{Nd}|_)*`
)

// tokenPatterns is a list of regular expression patterns that will be combined
// to match tokens.
//
// All subpatterns are non-capturing by using the syntax "(?: ...)" _except_
// for the following subpatterns that are named using the "(?P<name> ...)"
// syntax:
//
// - implicit (positional parameter with implicit position, i.e. "?")
// - explicit (positional parameter with explicit position, e.g. ":3")
// - named    (named parameter in ISO or MySQL style, e.g. ":foo" or "@foo")
// - python   (named parameter in python style, e.g. "%(foo)s")
//
// The named subpatterns are what we're after when matching tokens.  Anything
// else (even no match at all) is considered "other" and has .Kind==""
//
// The reason there are patterns other than those needed to capture the above
// is that we must identify tokens that may contain things that look like SQL
// parameters but that are not, such as comments and quoted strings.
var tokenPatterns = [...]string{
	// line comment
	// -- whatever until the end of the line (or the end of the file)
	`--[^\n]*(?:\n|$)`,

	// block comment
	// /* whatever until the matching */
	`/\*(?:[^*]|\*[^/])*\*/`,

	// single-quoted string
	// 'single-quoted string, maybe \'with\' escapes'
	`'(?:[^'\\]|\\.)*'`,

	// double-quoted string
	// "double-quoted string, maybe \"with\" escapes"
	`"(?:[^"\\]|\\.)*"`,

	// backtick string
	// `backtick string, maybe \`with\` escapes`
	"`(?:[^`\\\\]|\\\\.)*`",

	// implicit positional parameter
	// ?
	`(?P<implicit>\?)`,

	// explicit positional parameter
	// :4, :5, @1, @0 (zero is an invalid index, but is a valid Token)
	`[$@:](?P<explicit>` + natural + `)`,

	// named parameter
	// @userID, :name, %(python_style)s
	`[@:](?P<named>` + identifier + `)`,

	// python-style named parameter
	// %(foo)s, %(bar)s
	`%\((?P<python>` + identifier + `)\)s`}

var regexpMutex sync.Mutex
var compiledRegexp *regexp.Regexp

// tokenRegexp returns a pointer to a singleton instance of a compiled regular
// expression (compiledRegexp) used to match tokens.  It uses regexpMutex to
// prevent concurrent compilation of the regular expression.
func tokenRegexp() *regexp.Regexp {
	regexpMutex.Lock()
	defer regexpMutex.Unlock()

	if compiledRegexp != nil {
		return compiledRegexp
	}

	// It's not compiled yet, so we have to compile it.
	clauses := make([]string, len(tokenPatterns))
	for i, pattern := range tokenPatterns {
		// wrap the pattern so that it's a non-capturing subpattern
		clauses[i] = "(?:" + pattern + ")"
	}

	pattern := strings.Join(clauses, "|")
	compiledRegexp = regexp.MustCompile(pattern)

	return compiledRegexp
}

// Token is a chunk of a SQL query, possibly containing information about a
// SQL parameter therein.
type Token struct {
	// kind is one of the names from the named subpatterns in tokenPatterns,
	// e.g. "explicit", or "" if it didn't match a parameter-related pattern
	Kind string

	// text is the full extent of the Token in the source SQL, e.g. ":23"
	Text string

	// inside is the part of the Token relevant to interpretation, e.g. "foo"
	// in "@foo", or empty if no interpretation is necessary
	Inside string
}

// Lex returns a slice of tokens lexed (i.e. read, scanned) from query.  It is
// the opposite of Render.
func Lex(query string) []Token {
	var tokens = []Token{}
	regexp := tokenRegexp()
	// index of one-past-the-last-byte of the previous Token in query
	var previousTokenEnd = 0

	for _, match := range regexp.FindAllStringSubmatchIndex(query, -1) {
		begin, end := match[0], match[1]

		// If we skipped some text (no match in between), then consider the
		// skipped text to be a Token of other ("") kind.  For example, in:
		//
		//     /* here's a comment */ select * from foo where bar = ?;
		//
		// we will initially match the comment "/* here's a comment */" and
		// then match the implicit positional parameter "?".  The text in
		// between, " select * from foo where bar = ", is an other ("") Token.
		if begin != previousTokenEnd {
			tokens = append(tokens, Token{Text: query[previousTokenEnd:begin]})
		}

		// Determine which, if any, of the named subpatterns matched.  If none
		// matched, emit a Token of other ("") kind.
		submatchIndices := match[2:]
		currentToken := Token{Text: query[begin:end]}
		for i, subpatternName := range regexp.SubexpNames()[1:] {
			subBegin, subEnd := submatchIndices[2*i], submatchIndices[2*i+1]
			if subBegin == -1 {
				continue // this subpattern didn't match
			}

			currentToken.Kind = subpatternName
			currentToken.Inside = query[subBegin:subEnd]
			break // at most one subpattern will match (I claim)
		}

		tokens = append(tokens, currentToken)
		previousTokenEnd = end
	}

	if previousTokenEnd != len(query) {
		tokens = append(tokens, Token{Text: query[previousTokenEnd:]})
	}

	return tokens
}

// Render returns the concatenation of all of the text in tokens.  It is the
// opposite of Lex.
func Render(tokens []Token) string {
	texts := make([]string, len(tokens))
	for i, token := range tokens {
		texts[i] = token.Text
	}

	return strings.Join(texts, "")
}
