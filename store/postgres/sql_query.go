package postgres

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	// "github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"
)

type (
	Scanner = sql.Scanner
	Sqlizer = sq.Sqlizer

	SelectQ = sq.SelectBuilder
)

var (
	// PostgreSQL statement builder
	psql = sq.StatementBuilder.PlaceholderFormat(
		sq.Dollar,
	)
)

// sql.Scanner
type ScanFunc func(src any) error

var _ sql.Scanner = (ScanFunc)(nil)

// Scan implements database/sql.Scanner interface.
func (fn ScanFunc) Scan(src any) error {
	if fn != nil {
		return fn(src)
	}
	// ignore scan data ..
	return nil
}

// SET of UNIQUE name(s)
type names map[string]uint8

func (h names) exists(dn string) bool {
	_, ok := h[dn]
	return ok
}

func (h names) append(dn string) bool {
	if h.exists(dn) {
		return false
	}
	h[dn] = 0
	return true
}

func (h names) remove(dn string) {
	delete(h, dn)
}

// [C]ommon [T]able [E]xpression
type CTE struct {
	Name  string
	Cols  []string
	Query sq.Sqlizer // SQLizer
	// Materialized *bool // [ [ NOT ] MATERIALIZED ]
}

var _ Sqlizer = (*CTE)(nil)

func (e *CTE) ToSql() (CTE string, _ []any, err error) {
	query, args, err := e.Query.ToSql() // convertToSql(e.Source)
	if err != nil {
		return "", nil, err
	}
	CTE = e.Name
	if len(e.Cols) > 0 {
		CTE += "(" + strings.Join(e.Cols, ", ") + ")"
	}
	CTE += " AS"
	// if is := e.Materialized; is != nil {
	// 	if !(*is) {
	// 		CTE += " NOT"
	// 	}
	// 	CTE += " MATERIALIZED"
	// }
	return CTE + " (" + query + ")", args, nil
}

type WITH struct {
	Recursive bool
	common    []*CTE // ordered
	names     names  // hash[name]index
}

func (c *WITH) Has(name string) bool {
	dn := strings.ToLower(name)
	return c.names.exists(dn)
}

func (c *WITH) CTE(expr CTE) {
	name := expr.Name
	if c.Has(name) {
		panic(fmt.Errorf("WITH %q; -- DUPLICATE", name))
	}
	e := len(c.names)
	// if cte.Recursive && id > 0 {
	// 	panic(fmt.Errorf("WITH RECURSIVE %q; -- MUST be the first CTE!", name))
	// }
	if c.names == nil {
		c.names = make(names)
	}
	dn := strings.ToLower(name)
	c.names[dn] = uint8(e) // UP to 255 !
	c.common = append(c.common, &expr)
}

var _ Sqlizer = (*WITH)(nil)

func (c *WITH) ToSql() (WITH string, _ []any, err error) {
	var CTE string
	for e, view := range c.common {
		CTE, _, err = view.ToSql()
		if err != nil {
			return "", nil, err
		}
		if e > 0 {
			WITH += ", "
		} else {
			WITH = "WITH "
			if c.Recursive {
				WITH += " RECURSIVE "
			}
		}
		WITH += CTE
	}
	return WITH, nil, nil
}

type JOIN struct {
	Kind, // [INNER|CROSS|LEFT|RIGHT[ OUTER] ]JOIN
	Source, // RIGHT: [schema.]table(type)|LATERAL(SELECT)
	Alias, // AS
	Pred string // ON
}

var _ Sqlizer = (*JOIN)(nil)

func (e *JOIN) String() string {
	parts := make([]string, 2, 6)
	parts[0] = e.Kind
	parts[1] = e.Source
	if e.Alias != "" {
		parts = append(
			parts, "AS", e.Alias,
		)
	}
	parts = append(
		parts, "ON", coalesce(e.Pred, "true"),
	)
	return strings.Join(parts, " ")
}

func (rel *JOIN) ToSql() (join string, _ []interface{}, err error) {
	return rel.String(), nil, nil
}

type statement struct {
	// WITH clause ..
	WITH
	// Query expression
	Query sq.Sqlizer // sq.SelectBuilder
	// Parameters; Named
	Params map[string]any
	// map[schema.table]cte.name
	Resource map[string]string
}

func (q *statement) Sql() (SQL string, err error) {
	SQL, _, err = q.Query.ToSql()
	if err != nil {
		SQL = ""
		return // "", nil, err
	}
	var WITH string
	WITH, _, err = q.WITH.ToSql()
	if err != nil {
		SQL = ""
		return // "", err
	}
	SQL = WITH + SQL // AS prefix
	return           // SQL, nil
	// var (
	// 	WITH   string
	// 	SELECT = q.Query.Suffix("") // shallowcopy
	// )
	// WITH, _, err = q.WITH.ToSql()
	// if err != nil {
	// 	return // "", nil, err
	// }
	// if WITH != "" {
	// 	SELECT = SELECT.Prefix(WITH)
	// }
	// query, _, err = SELECT.ToSql()
	// return
}

func (q *statement) ToSql() (query string, args []any, err error) {
	query, err = q.Sql()
	if err == nil && len(q.Params) > 0 {
		query, args, err = BindNamed(query, q.Params)
	}
	return // query, args, err
}

// CompactSQL formats given SQL text to compact form.
// - replaces consecutive white-space(s) with single SPACE(' ')
// - suppress single-line comment(s), started with -- up to [E]nd[o]f[L]ine
// - suppress multi-line comment(s), enclosed into /* ... */ pairs
// - transmits literal '..' or ".." sources in their original form
// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-OPERATORS
func CompactSQL(s string) string {

	var (
		r = strings.NewReader(s)
		w strings.Builder
	)

	w.Grow(int(r.Size()))

	var (
		err  error
		char rune
		last rune
		hold rune

		isSpace = func() (is bool) {
			switch char {
			case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
				is = true
			}
			return // false
		}
		isPunct = func(char rune) (is bool) {
			switch char {
			// none; start of text
			case 0:
				is = true
			// special
			// ':' USES [squirrel] for :named parameters,
			//     so we need to keep SPACE if there were any
			case ',', '(', ')', '[', ']', ';', '\'', '"': // , ':':
				is = true
			// operators
			case '+', '-', '*', '/', '<', '>', '=', '~', '!', '@', '#', '%', '^', '&', '|':
				is = true
			}
			return // false
		}
		isQuote = func() (is bool) {
			switch char {
			case '\'', '"': // SQUOTE, DQUOTE:
				is = true
			}
			return // false
		}
		// context
		space   bool // [IN] [w]hite[sp]ace(s)
		quote   rune // [IN] [l]i[t]eral(s); *QUOTE(s)
		comment rune // [IN] [c]o[m]ment; [-|*]
		// helpers
		isComment = func() bool {
			switch comment {
			case '-':
				{
					// comment: close(\n)
					if char == '\n' { // EOL
						space = true // inject
						comment = 0  // close
						hold = 0     // clear
					}
					return true // still IN ...
				}
			case '*':
				{
					// comment: close(*/)
					if hold == 0 && char == '*' {
						// MAY: close(*/)
						hold = char
						// need more data ...
					} else if hold == '*' && char == '/' {
						space = true // inject
						comment = 0  // close
						hold = 0     // clear
					}
					return true // still IN ...
				}
				// default: 0
			}
			// NOTE: (comment == 0)
			switch hold {
			// comment: start(--)
			case '-': // single-line
				{
					if char == hold {
						hold = 0       // clear
						comment = char // start
						return true
					}
					return false
				}
			// comment: start(/*)
			case '/': // multi-line
				{
					if char == '*' {
						hold = 0       // clear
						comment = char // start
						return true
					}
					return false
				}
			case 0:
				{
					// NOTE: (hold == 0)
					switch char {
					case '-':
					case '/':
					default:
						// NOT alike ...
						return false
					}
					// need more data ...
					hold = char
					// DO NOT write(!)
					return true
				}
			default:
				{
					// NO match
					// need to write hold[ed] char
					return false
				}
			}
		}
		isLiteral = func() bool {
			if !isQuote() || last == '\\' { // ESC(\')
				return quote > 0 // We are IN ?
			}
			// close(?)
			if quote == char { // inLiteral(?)
				quote = 0
				return true // as last
			}
			// start(!)
			quote = char
			return true
		}
		// [re]write
		output = func() {
			if hold > 0 {
				w.WriteRune(hold)
				last = hold
				hold = 0
			}
			if space {
				space = false
				if !isPunct(last) && !isPunct(char) {
					w.WriteRune(' ') // INJECT SPACE(' ')
				}
			}
			w.WriteRune(char)
			last = char
		}
	)

	var e int
	for {

		char, _, err = r.ReadRune()
		if err != nil {
			break
		}
		e++ // char index position

		if isComment() {
			// suppress; DO NOT write(!)
			continue
		}

		if isLiteral() {
			// [re]write: as is (!)
			output()
			continue
		}

		if isSpace() {
			// fold sequence ...
			space = true
			continue
		}
		// [re]write: [hold]char
		output()
	}

	if err != io.EOF {
		panic(err)
	}

	return w.String()
}
