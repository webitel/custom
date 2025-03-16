package postgres

import (
	"fmt"
	"strconv"
	"unicode"

	"errors"
	// "github.com/pkg/errors"
	// "github.com/jmoiron/sqlx/reflectx"
	// "github.com/jmoiron/sqlx"
)

// Parameters as [named] values
type Parameters map[string]any

func (c Parameters) Get(name string) (value any, ok bool) {
	if c != nil {
		value, ok = c[name]
	}
	return // value, ok
}

func (c Parameters) Set(name string, value any) (ok bool) {
	if name == "" {
		return // false
	}
	c[name] = value
	ok = true
	return // ok
}

func (c Parameters) Add(name string, value any) (ok bool) {
	if name == "" {
		return // false
	}
	if _, ok = c.Get(name); ok {
		ok = false // exists !
		panic(fmt.Errorf("params.add(%s); duplicate", name))
	}
	c[name] = value
	ok = true
	return // ok
}

func (c Parameters) Del(name string) (ok bool) {
	if _, ok = c.Get(name); ok {
		delete(c, name)
	}
	return // ok ?
}

// -- Compilation of Named Queries

// Bindvar types supported by Rebind, BindMap and BindStruct.
const (
	UNKNOWN = iota
	QUESTION
	DOLLAR
	NAMED
	AT
)

// Allow digits and letters in bind params;  additionally runes are
// checked against underscores, meaning that bind params can have be
// alphanumeric with underscores.  Mind the difference between unicode
// digits and numbers, where '5' is a digit but '五' is not.
var allowedBindRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

// FIXME: this function isn't safe for unicode named params, as a failing test
// can testify.  This is not a regression but a failure of the original code
// as well.  It should be modified to range over runes in a string rather than
// bytes, even though this is less convenient and slower.  Hopefully the
// addition of the prepared NamedStmt (which will only do this once) will make
// up for the slightly slower ad-hoc NamedExec/NamedQuery.

// compile a NamedQuery into an unbound query (using the '?' bindvar) and
// a list of names.
func compileNamedQuery(qs []byte, bindType int, offset int) (query string, names []string, err error) {
	names = make([]string, 0, 10)
	rebound := make([]byte, 0, len(qs))

	inName := false
	last := len(qs) - 1
	name := make([]byte, 0, 10)
	currentVar := 1
	if offset < 0 {
		offset = 0
	}

	for i, b := range qs {
		// a ':' while we're in a name is an error
		if b == ':' {
			// if this is the second ':' in a '::' escape sequence, append a ':'
			if inName && i > 0 && qs[i-1] == ':' {
				rebound = append(rebound, ':')
				inName = false
				continue
			} else if inName {
				err = errors.New("unexpected `:` while reading named param at " + strconv.Itoa(i))
				return query, names, err
			}
			inName = true
			name = []byte{}
		} else if inName && i > 0 && b == '=' && len(name) == 0 {
			rebound = append(rebound, ':', '=')
			inName = false
			continue
			// if we're in a name, and this is an allowed character, continue
		} else if inName && (unicode.IsOneOf(allowedBindRunes, rune(b)) || b == '_' || b == '.') && i != last {
			// append the byte to the name if we are in a name and not on the last byte
			name = append(name, b)
			// if we're in a name and it's not an allowed character, the name is done
		} else if inName {
			inName = false
			// if this is the final byte of the string and it is part of the name, then
			// make sure to add it to the name
			if i == last && unicode.IsOneOf(allowedBindRunes, rune(b)) {
				name = append(name, b)
			}
			// add the string representation to the names list
			names = append(names, string(name))
			// add a proper bindvar for the bindType
			switch bindType {
			// oracle only supports named type bind vars even for positional
			case NAMED:
				rebound = append(rebound, ':')
				rebound = append(rebound, name...)
			case QUESTION, UNKNOWN:
				rebound = append(rebound, '?')
			case DOLLAR:
				// TODO: find named position in list ?
				param := string(name)
				for p, _ := range names {
					// MUST: be at least as last !
					if param == names[p] {
						// index as position
						currentVar = (p) + 1
						// param: already bound ?
						// are we at the middle of the list ?
						if currentVar != len(names) {
							// duplicate: remove last !
							names = names[:len(names)-1]
						}
						break
					}
				}
				rebound = append(rebound, '$')
				for _, b := range strconv.Itoa(currentVar + offset) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			case AT:
				rebound = append(rebound, '@', 'p')
				for _, b := range strconv.Itoa(currentVar) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			}
			// add this byte to string unless it was not part of the name
			if i != last {
				rebound = append(rebound, b)
			} else if !unicode.IsOneOf(allowedBindRunes, rune(b)) {
				rebound = append(rebound, b)
			}
		} else {
			// this is a normal byte and should just go onto the rebound query
			rebound = append(rebound, b)
		}
	}

	return string(rebound), names, err
}

// func bindNamedMapper(bindType int, query string, arg interface{}, m *reflectx.Mapper) (string, []interface{}, error) {
// 	if maparg, ok := arg.(map[string]interface{}); ok {
// 		return bindMap(bindType, query, maparg)
// 	}
// 	switch reflect.TypeOf(arg).Kind() {
// 	case reflect.Array, reflect.Slice:
// 		return bindArray(bindType, query, arg, m)
// 	default:
// 		return bindStruct(bindType, query, arg, m)
// 	}
// }

// like bindArgs, but for maps.
func bindMapArgs(names []string, arg map[string]any) ([]any, error) {
	arglist := make([]any, 0, len(names))

	for _, name := range names {
		val, ok := arg[name]
		if !ok {
			return arglist, fmt.Errorf("could not find name %s in %#v", name, arg)
		}
		arglist = append(arglist, val)
	}
	return arglist, nil
}

// bindMap binds a named parameter query with a map of arguments.
func bindMap(bindType int, query string, args map[string]any, offset int) (string, []any, error) {
	bound, names, err := compileNamedQuery([]byte(query), bindType, offset)
	if err != nil {
		return "", []any{}, err
	}

	arglist, err := bindMapArgs(names, args)
	return bound, arglist, err
}

// QUERY: ("select :p, :p, :p", p=1)
// NAMED: ("select $1, $2, $3", $1=1, $2=1, $3=1)
// TODO:  ("select $1, $1, $1", $1=1)
func BindNamed(query string, params map[string]any) (string, []any, error) {
	// // return bindNamedMapper(BindType(db.driverName), query, arg, db.Mapper)
	// return sqlx.BindNamed(sqlx.DOLLAR, query, params)
	return bindMap(DOLLAR, query, params, 0)
}

// QUERY: ("select :p, :p, :p", p=1)
// NAMED: ("select $1, $2, $3", $1=1, $2=1, $3=1)
// TODO:  ("select $1, $1, $1", $1=1)
func BindNamedOffset(query string, params map[string]any, offset int) (string, []any, error) {
	// // return bindNamedMapper(BindType(db.driverName), query, arg, db.Mapper)
	// return sqlx.BindNamed(sqlx.DOLLAR, query, params)
	return bindMap(DOLLAR, query, params, offset)
}

// func BindNamedAs(bindType int, query string, params map[string]any) (string, []any, error) {
// 	// // return bindNamedMapper(BindType(db.driverName), query, arg, db.Mapper)
// 	// return sqlx.BindNamed(sqlx.DOLLAR, query, params)
// 	return bindMap(bindType, query, params)
// }
