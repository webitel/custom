//

package postgres

import (
	"fmt"
	"path"
	"strings"

	custom "github.com/webitel/custom/data"
	"github.com/webitel/custom/store"
)

type datasetOptions struct {
	// Request options
	Req store.SearchOptions
	// [D]omain [C]omponent ID
	Dc int64
	// View as a domain scope
	View struct {
		// Name of the view
		// e.g.: "domain", "custom", "system"
		Name string
		//
		expr string
	}
	// Additional filter assertion(s)
	Dir        *string // path
	Name       string  // path
	Title      string
	Readonly   *bool
	Extendable *bool
}

func newDatasetOptions(opts []store.SearchOption) datasetOptions {
	opt := datasetOptions{
		Req: store.NewSearch(opts...),
	}
	// Zero(-based) ; (-0) - system ; (1+) - custom
	opt.Dc = max(opt.Req.Dc, 0)

	const (
		pdc  = "pdc"
		left = aliasType
	)
	var (
		rdc   = sqlident{left, columnDc}
		vbool custom.BoolValue
		vtext custom.StringValue
	)
	// opt.Params.Set(pdc, opt.Dc) // Zero(-based) !

	// ALL( system + custom )
	opt.View.Name = "domain"
	opt.View.expr = fmt.Sprintf(
		"COALESCE(%[1]s,:%[2]s)=:%[2]s",
		rdc, pdc,
	)
	for name, assert := range opt.Req.Filter {
		switch name {
		// base [dir] of the path
		case "dir":
			{
				// dir, _ := assert.(string)
				// opt.Dir = &dir // Allow: ""
				err := vtext.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect string value", name, assert))
				}
				opt.Dir = vtext.Interface().(*string)
			}
		// relative [path] match
		case "path":
			{
				err := vtext.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect string value", name, assert))
				}
				vs := vtext.Interface().(*string)
				if vs == nil {
					continue
				}
				s := (*vs)
				// s, _ := assert.(string)
				if customFilterIsPresent(s) {
					continue // skip ; all records has [path] assigned
				}
				dir, name := path.Split(s)
				if dir != "" && opt.Dir == nil {
					opt.Dir = &dir
				}
				opt.Name = name
			}
		// path filename
		case "repo", "id":
			{
				err := vtext.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect string value", name, assert))
				}
				vs := vtext.Interface().(*string)
				if vs == nil {
					continue
				}
				opt.Name = (*vs)
				// s, _ := assert.(string)
				// opt.Name = s
			}
		// title ; lang specific
		case "name", "title":
			{
				err := vtext.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect string value", name, assert))
				}
				vs := vtext.Interface().(*string)
				if vs == nil {
					continue
				}
				opt.Title = (*vs)
				// s, _ := assert.(string)
				// opt.Title = s
			}
		// [NOT] GLOBAL ?
		case "readonly":
			{
				err := vbool.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect boolean value", name, assert))
				}
				opt.Readonly = vbool.Interface().(*bool)
				// if is, _ := assert.(bool); is {
				// 	// force "global" types ONLY !
				// 	// ignore "pdc" param value !
				// 	opt.View.Name = "system"
				// 	opt.View.expr = rdc.String() + " ISNULL"
				// } else {
				// 	// force "custom" types ONLY !
				// 	// ctx.params["pdc"] = domain.pdc // MAYBE: 0 !
				// 	opt.View.Name = "custom"
				// 	opt.View.expr = rdc.String() + " = :pdc" // Zero(0) - will be invalidate !
				// }
			}
		// [NOT] GLOBAL & extendable ?
		case "extendable":
			{
				err := vbool.Decode(assert)
				if err != nil {
					panic(fmt.Errorf("dataset[%s] = %T ; expect boolean value", name, assert))
				}
				opt.Extendable = vbool.Interface().(*bool)
				// if is, _ := assert.(bool); is {
				// 	// force "global" types ONLY !
				// 	// ignore "pdc" param value !
				// 	opt.View.Name = "system"
				// 	opt.View.expr = rdc.String() + " ISNULL"
				// 	// .. AND t.extendable
				// } else {
				// 	// .. NOT t.extendable
				// }
			}
		case "available":
			// LEFT JOIN pg_catalog.class
		}
	}
	return opt
}

func (c *datasetOptions) Apply(query SelectQ, params Parameters) SelectQ {
	// Filter View moderation(s)
	const (
		pdc  = "pdc"
		left = aliasType
	)
	var (
		rdc = sqlident{left, columnDc}
	)
	if c.Readonly != nil {
		if is := (*c.Readonly); is {
			// force "global" types ONLY !
			// ignore "pdc" param value !
			c.View.Name = "system"
			c.View.expr = rdc.String() + " ISNULL"
		} else {
			// force "custom" types ONLY !
			// ctx.params["pdc"] = domain.pdc // MAYBE: 0 !
			c.View.Name = "custom"
			c.View.expr = rdc.String() + " = :pdc" // Zero(0) - will be invalidate !
		}
	}
	if c.Extendable != nil {
		if is := (*c.Extendable); is {
			// "GLOBAL" types CAN be [extendable] ONLY !
			// ignore "pdc" param value !
			c.View.Name = "system"
			c.View.expr = rdc.String() + " ISNULL"
		} else {
			// no affect !
		}
	}
	// Mandatory [pdc] filter !
	params.Set(pdc, c.Dc)
	query = query.Where(c.View.expr)

	for _, filter := range []func(*datasetOptions, SelectQ, Parameters) SelectQ{
		// whereDatasetReadonly,
		whereDatasetExtendable,
		//
		whereDatasetDir,
		whereDatasetName,
		whereDatasetTitle,
	} {
		query = filter(c, query, params)
	}
	return query
}

func whereDatasetDir(where *datasetOptions, query SelectQ, params Parameters) SelectQ {
	if where.Dir == nil {
		return query
	}
	dir := (*where.Dir)
	dir = strings.Trim(dir, "/")
	match, assert := "=", strings.ToLower(dir)
	if customFilterIsSubstring(dir) {
		match, assert = "ILIKE", customFilterSubstringAssertion(dir)
	}
	const (
		param = "dir"
		left  = aliasType
	)
	params.Add(param, assert)
	query = query.Where(fmt.Sprintf(
		"COALESCE(%s,'') %s :%s",
		sqlident{left, columnTypeDir}, match, param,
	))
	return query
}

func whereDatasetName(where *datasetOptions, query SelectQ, params Parameters) SelectQ {
	name := where.Name
	if customFilterIsPresent(name) {
		return query // skip ; all records has [name] assigned !
	}
	match, assert := "=", strings.ToLower(name)
	if customFilterIsSubstring(name) {
		match, assert = "ILIKE", customFilterSubstringAssertion(name)
	}
	const (
		param = "name"
		left  = aliasType
	)
	params.Add(param, assert)
	query = query.Where(fmt.Sprintf(
		"%s %s :%s",
		sqlident{left, columnTypeName}, match, param,
	))
	return query
}

func whereDatasetTitle(where *datasetOptions, query SelectQ, params Parameters) SelectQ {
	title := where.Title
	if customFilterIsPresent(title) {
		return query // skip ; all records has [title] assigned !
	}
	match, assert := "=", title
	if customFilterIsSubstring(title) {
		match, assert = "ILIKE", customFilterSubstringAssertion(title)
	}
	const (
		param = "title"
		left  = aliasType
	)
	params.Add(param, assert)
	query = query.Where(fmt.Sprintf(
		"COALESCE(%[1]s.%[2]s.,(%[1]s.%[3]s)::::text) %s :%s COLLATE \"default\"",
		left, columnTypeTitle, columnTypeName, match, param,
	))
	return query
}

func whereDatasetReadonly(where *datasetOptions, query SelectQ, _ Parameters) SelectQ {
	if where.Readonly == nil {
		return query
	}
	expr := "ISNULL"
	if !(*where.Readonly) {
		expr = "NOTNULL"
	}
	query = query.Where(fmt.Sprintf(
		"%s %s",
		sqlident{aliasType, columnDc}, expr,
	))
	return query
}

func whereDatasetExtendable(where *datasetOptions, query SelectQ, _ Parameters) SelectQ {
	if where.Extendable == nil {
		return query
	}
	expr := "" // IS
	if !(*where.Extendable) {
		expr = "NOT "
	}
	query = query.Where(fmt.Sprintf(
		"%s(%s)",
		expr, sqlident{aliasType, columnTypeExtendable},
	))
	return query
}
