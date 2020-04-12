namedsql
========
`select @this in @that` → `select ? in (?, ?, ?)`

Why
---
> What's in a name? That which we call a rose <br/>
> By any other name would smell as sweet

What
----
`namedsql` is a Go package for compiling SQL into SQL.  The input language is
SQL containing named parameters, possibly including to array-like values, and
the output language is SQL containing only positional parameters (i.e. `?`),
and no array-valued parameters.

How
---
```Go
import (
	"database/sql"

	"github.com/dgoffredo/namedsql"
)

type UserID = uint64

const (
	gender      = 0
	orientation = 1
)

func GetTags(userID UserID, execute func(string, ...interface{}) (*sql.Rows, error)) ([]int, error) {
	// use named parameters and parameters that are slices (or arrays)
	query, bindings := namedsql.MustArrangeAndExpand(`
         select value
         from tags
         where type in @types
           and userid = @userID`,
		map[string]interface{}{
			"types":  []int{gender, orientation},
			"userID": userID})

	// now `query` contains only "?" parameters, and `bindings` is a slice
	rows, err := execute(query, bindings...)

	if err != nil {
		return nil, err
	}

	// unpack the results
	values := make([]int)
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}
```

More
----
### `Arrange(query, bindings, more...)`
takes a SQL `query`, a name-keyed `map` of `bindings`, and optionally `more`
positional bindings, and returns a tuple `(newQuery, positionalBindings, err)`,
where `newQuery` is a version of `query` containing only positional
parameters, and `positionalBindings` is a slice of parameters based in
`bindings` and `more`, but in the correct order for use with `newQuery`.

If an error occurs, then both `newQuery` and `positionalBindings` are
default values and `err` is not `nil.`

For example,
```Go
import "github.com/dgoffredo/namedsql"

id := 1337
color := "purple"
limit := 10

query, bindings, err := namedsql.Arrange(
	"select * from llamas where id=? or friend=:1 and color=:color limit :limit;",
	map[string]interface{}{"color": color, "limit": limit},
	id)
```
leaves `query` with the value
```sql
select * from llamas where id=? or friends=? and color=? limit ?;
```
and `bindings` with the value
```Go
[]interface{}{id, id, color, limit}
```

### `Expand(query, positionalBindings...)`
takes a SQL `query` and `positionalBindings` interfaces, and returns a tuple
`(newQuery, newBindings, err)`, where `newQuery` is a version of `query`
that has had its array-valued parameters "exploded" into lists of
per-element parameters, and `newBindings` is the corresponding "exploded"
bindings.

If an error occurs, then both the returned `newQuery` and `positionalBindings`
are default values and `err` is not `nil.`

For example,
```Go
import "github.com/dgoffredo/namedsql"

names := []string{"Moe", "Larry", "Curly"}
limit := 3
query, bindings, err := namedsql.Expand(
    "select * from stooges where name in ? limit ?",
    names,
    limit)
```
leaves `query` with the value
```sql
select * from stooges where name in (?, ?, ?) limit ?
```
and `bindings` with the value
```Go
[]interface{}{names[0], names[1], names[2], limit}
```

### `ArrangeAndExpand(query, bindings, more...)`
performs `Arrange` followed by `Expand`, but parsing the SQL only once.

### `MustArrange`, `MustExpand`, and `MustArrangeAndExpand`
are variants of the above functions, but that rather than returning a trailing
`error` result, instead panic on failure.  These make sense to use in contexts
where the input query and the names and number of bindings are hard-coded.

Parameter Language
------------------
Any of the following are supported:
- `?` is a positional parameter
- `$n` is an explicit positional parameter for some positive integer `n`.
  It refers to the `n`'th positional parameter (one-based).
- `:n`, as above.
- `@n`, as above.
- `:identifier` is a named parameter for some identifier `identifier`, where an
  identifier is a combination of Unicode letters, digits, and underscores, but
  not beginning with a digit.
- `@identifier`, as above.
- `%(identifier)s`, as above.

The SQL query output by `Arrange` will contain only `?`-style parameters.

Note that the parameter name `identifier` cannot be enclosed in quotes &mdash;
not even backticks.  It simplifies things.
