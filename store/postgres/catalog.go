package postgres

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	custom "github.com/webitel/custom/data"
	"github.com/webitel/custom/store"
	custompb "github.com/webitel/proto/gen/custom"
	datatypb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type Catalog struct {
	cluster
}

func NewCatalog(dc ...*pgxpool.Pool) *Catalog {
	return &Catalog{
		cluster: cluster{
			dc: dc,
		},
	}
}

var _ store.Catalog = (*Catalog)(nil)

func (c *Catalog) Search(opts ...store.SearchOption) (*custompb.DatasetList, error) {

	ctx, err := customDatasetQuery(opts)
	if err != nil {
		return nil, err
	}

	query, args, err := ctx.ToSql()
	if err != nil {
		return nil, err
	}

	dc := c.secondary()
	rows, err := dc.Query(ctx.req.Context, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var page custompb.DatasetList
	err = customDatasetFetchRows(ctx, rows, &page)
	if err != nil {
		return nil, customSchemaError(err)
	}
	return &page, nil
}

type datasetQ struct {
	datasetOptions
	*query[*custompb.Dataset]
	tables map[string]string
	joined names
}

func customDatasetQuery(opts []store.SearchOption) (*datasetQ, error) {
	cte := &datasetQ{
		datasetOptions: newDatasetOptions(opts),
		query:          newQuery[*custompb.Dataset](),
		tables:         make(map[string]string),
		joined:         make(names),
	}
	cte.req = store.NewSearch(opts...)
	err := customDatasetSelectQuery(cte)
	return cte, err
}

func customDatasetSelectQuery(ctx *datasetQ) error {

	table := sqlident{schemaCustom, tableType}
	if rel, is := ctx.Resource[table.String()]; is {
		table = []string{rel}
	}

	const (
		left = aliasType
	)

	query := psql.Select().From(strings.Join(
		[]string{table.String(), left}, " ",
	)).Columns(
		// sqlident{left, colmnDc}.String(),
		// sqlident{left, colmnId}.String(),
		// [name]
		sqlident{left, columnTypeName}.String(),
		// [readonly]
		fmt.Sprintf("(%s ISNULL)",
			sqlident{left, columnDc},
		),
	)

	params := ctx.Params
	if params == nil {
		params = make(map[string]any)
		ctx.Params = params
	}

	var (
		cols = names{} // columns

		// values
		vbool pgtype.Bool
		vtext pgtype.Text
		vtime pgtype.Timestamp
		// plan data decoding
		plan = dataScanPlan[*custompb.Dataset]{
			// [name]
			func(row *custompb.Dataset) sql.Scanner {
				return ScanFunc(func(src any) error {
					// decode
					err := vtext.Scan(src)
					if err != nil {
						return err
					}
					// row.Repo
					row.Repo = vtext.String
					return nil
				})
			},
			// [readonly]
			func(row *custompb.Dataset) sql.Scanner {
				return ScanFunc(func(src any) error {
					// decode
					err := vbool.Scan(src)
					if err != nil {
						return err
					}
					row.Readonly = vbool.Bool
					return nil
				})
			},
		}
	)

	var (
		rels = names{} // relations ; joins
		// joinAcl = func() {
		// 	const alias = "acl"
		// 	if !rels.append(alias) {
		// 		return // duplicate !
		// 	}
		// }
		joinRole = func(fkey, alias string) bool {
			if !rels.append(alias) {
				return false // duplicate !
			}
			query = query.JoinClause(fmt.Sprintf(
				"LEFT JOIN %s %s ON %s.%s = %[2]s.id",
				sqlident{schemaDir, tableAuth}, alias, left, fkey,
			))
			return true
		}
		viewRole = func(fkey, alias string) error {
			if !cols.append(fkey) {
				return nil // duplicate !
			}
			// ensure joined
			_ = joinRole(fkey, alias)
			var csb strings.Builder
			fmt.Fprintf(&csb, "(SELECT ROW(%s.id", alias)
			for _, fd := range []string{"id", "name"} {
				switch fd {
				case "id":
					// default
				case "name":
					fmt.Fprintf(&csb, ",coalesce(%s.name,(%[1]s.auth)::::text)", alias)
				case "type":
					// const "users"
				default:
					return custom.RequestError(
						"custom.lookup.field.unknown",
						"custom: dataset{%s{%s}} no such field",
						fkey, fd,
					)
				}
			}
			// // fmt.Fprintf(&csb, ") %s", alias)
			// csb.WriteByte(')')
			fmt.Fprintf(&csb, ") WHERE %s.id NOTNULL)", alias) // global (system) types has no owner(author) !
			query = query.Column(
				csb.String(),
			)
			return nil
		}
		// joinAuthor = func() string {
		// 	const alias = "own"
		// 	joinRole(columnCreatedBy, alias)
		// 	return alias
		// }
		// joinEditor = func() string {
		// 	const alias = "edt"
		// 	joinRole(columnUpdatedBy, alias)
		// 	return alias
		// }
	)

	// SELECT column(s)..
	var (
		all         bool
		fields      = make(names, 16)
		operational = []string{
			"repo", "path",
			"name", "about",
			"fields", "primary", "display",
			"indices",
			"readonly", "extendable",
			"created_at", "created_by",
			"updated_at", "updated_by",
		}
	)
	for _, name := range ctx.req.Fields {
		name = strings.ToLower(name)
		switch name {
		case "+", "*":
			if all {
				continue // once
			}
			all = true // once
			for _, name := range operational {
				fields.append(name)
			}
			continue // for
		}
		// requested
		fields.append(name)
	}
	for name := range fields {
		switch name {
		case "id", "repo":
			// core ; always !
		case "path":
			{
				const column = columnTypePath
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(fmt.Sprintf(
					"COALESCE((%[1]s.%[2]s||'/'),'')||%[1]s.%[3]s",
					left, columnTypeDir, columnTypeName, // path: [dir/]scope", sqlident{left, column},
				))
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.Path = ""
						err := vtext.Scan(src)
						if err != nil {
							return err
						}
						row.Path = vtext.String
						return nil
					})
				})
			}
		case "name", "title":
			{
				const column = columnTypeTitle
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.Name = ""
						err := vtext.Scan(src)
						if err != nil {
							return err
						}
						row.Name = vtext.String
						return nil
					})
				})
			}
		case "about", "usage":
			{
				const column = columnTypeUsage
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.About = ""
						err := vtext.Scan(src)
						if err != nil {
							return err
						}
						row.About = vtext.String
						return nil
					})
				})
			}
		case "fields":
			{
				const column = "fields"
				if !cols.append(column) {
					break // duplicate !
				}
				// var (
				// 	fetch  customTypeScan
				// 	fieldQ = graph.Query{
				// 		Name: "fields",
				// 		Args: map[string]any{
				// 			"size": -1,
				// 		},
				// 		Fields: []*graph.Query{},
				// 	}
				// )
				ctx.Query = query
				ctx.plan = plan
				err := customDatasetFieldQuery(
					ctx, ctx.Resource[sqlident{schemaCustom, tableField}.String()], column,
				)
				if err != nil {
					return err
				}
				query = ctx.Query.(sq.SelectBuilder)
				plan = ctx.plan
				// ctx.plan = append(ctx.plan, fetch)
			}
		case "primary":
			{
				const column = columnTypeFieldPrimary
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.Primary = ""
						err := vtext.Scan(src)
						if err != nil {
							return err
						}
						row.Primary = vtext.String
						return nil
					})
				})
			}
		case "display":
			{
				const column = columnTypeFieldDisplay
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.Display = ""
						err := vtext.Scan(src)
						if err != nil {
							return err
						}
						row.Display = vtext.String
						return nil
					})
				})
			}
		case "indices":
			{
				// TODO:
			}
		case "readonly":
			// core ; always ! [dc] related
		case "extendable":
			{
				const column = columnTypeExtendable
				if !cols.append(column) {
					break // duplicate !
				}
				// READONLY -AND- extendable
				query = query.Column(fmt.Sprintf(
					"(%[1]s.%[2]s ISNULL AND %[1]s.%[3]s)",
					left, columnDc, column,
				))
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						err := vbool.Scan(src)
						if err != nil {
							return err
						}
						row.Extendable = vbool.Bool
						return nil
					})
				})
			}
		case "created_at":
			{
				const column = columnCreatedAt
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.CreatedAt = 0
						err := vtime.Scan(src)
						if err != nil {
							return err
						}
						if vtime.Valid {
							row.CreatedAt = vtime.Time.UnixMilli()
						}
						return nil
					})
				})
			}
		case "updated_at":
			{
				const column = columnUpdatedAt
				if !cols.append(column) {
					break // duplicate !
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return ScanFunc(func(src any) error {
						// decode
						row.UpdatedAt = 0
						err := vtime.Scan(src)
						if err != nil {
							return err
						}
						if vtime.Valid {
							row.UpdatedAt = vtime.Time.UnixMilli()
						}
						return nil
					})
				})
			}
		case "created_by":
			{
				const (
					alias  = "own"
					column = columnCreatedBy
				)
				// if !cols.append(column) {
				// 	break // duplicate !
				// }
				err := viewRole(column, alias)
				if err != nil {
					return err
				}
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return customScanLookup(&row.CreatedBy, false)
				})
			}
		case "updated_by":
			{
				const (
					alias  = "edt"
					column = columnUpdatedBy
				)
				// if !cols.append(column) {
				// 	break // duplicate !
				// }
				err := viewRole(column, alias)
				if err != nil {
					return err
				}
				plan = append(plan, func(row *custompb.Dataset) sql.Scanner {
					return customScanLookup(&row.UpdatedBy, false)
				})
			}
		default:
			{
				// unknown field name
				err := custom.RequestError(
					"custom.dataset.field.invalid",
					"custom: dataset{%s} no such field",
					name,
				)
				return err
			}
		}
	}

	// --------------------------------- //
	// [SELECT] filter(s)..
	// --------------------------------- //
	// // [P]rimary [D]omain [C]omponent ID
	// pdc := max(ctx.req.Dc, 0)
	// rdc := sqlident{left, columnDc}
	// view := struct {
	// 	name string
	// 	expr string
	// }{
	// 	name: "sys", // GLOBAL ONLY !
	// 	expr: rdc.String() + " ISNULL",
	// }
	// if pdc > 0 {
	// 	// ctx.params["pdc"] = domain.pdc // int8
	// 	view.name = "all"
	// 	view.expr = fmt.Sprintf(
	// 		"COALESCE(%s,:pdc)=:pdc",
	// 		rdc,
	// 	)
	// } else {
	// 	pdc = 0 // Global ONLY !
	// 	// ctx.params["pdc"] = (*int64)(nil)
	// 	// view.expr = ident(from, "dc") + " ISNULL",
	// }
	// var (
	// 	// for duplicates invalidation
	// 	where    = names{}
	// 	whereDir = func(s string) bool {
	// 		// if customFilterIsPresent(s) {}
	// 		if s == "*" {
	// 			return false // skip ; all records has [path] assigned
	// 		}
	// 		const name = "dir"
	// 		if !where.append(name) {
	// 			return false // duplicate
	// 		}
	// 		match, assert := "=", strings.ToLower(s)
	// 		if customFilterIsSubstring(s) {
	// 			match, assert = "ILIKE", customFilterSubstringAssertion(s)
	// 		}
	// 		params[name] = assert
	// 		query = query.Where(fmt.Sprintf(
	// 			"COALESCE(%s,'') %s :%s",
	// 			sqlident{left, columnTypeDir}, match, name,
	// 		))
	// 		return true
	// 	}
	// 	whereName = func(s string) bool {
	// 		if customFilterIsPresent(s) {
	// 			return false // skip ; all records has [path] assigned
	// 		}
	// 		const name = "name"
	// 		if !where.append(name) {
	// 			return false // duplicate
	// 		}
	// 		match, assert := "=", strings.ToLower(s)
	// 		if customFilterIsSubstring(s) {
	// 			match, assert = "ILIKE", customFilterSubstringAssertion(s)
	// 		}
	// 		params[name] = assert
	// 		query = query.Where(fmt.Sprintf(
	// 			"%s %s :%s",
	// 			sqlident{left, columnTypeName}, match, name,
	// 		))
	// 		return true
	// 	}
	// 	whereTitle = func(s string) bool {
	// 		if customFilterIsPresent(s) {
	// 			return false // skip ; all records has [title] assigned
	// 		}
	// 		const name = "title"
	// 		if !where.append(name) {
	// 			return false // duplicate
	// 		}
	// 		match, assert := "=", s
	// 		if customFilterIsSubstring(s) {
	// 			match, assert = "ILIKE", customFilterSubstringAssertion(s)
	// 		}
	// 		params[name] = assert
	// 		query = query.Where(fmt.Sprintf(
	// 			"%s %s :%s",
	// 			sqlident{left, columnTypeTitle}, match, name,
	// 		))
	// 		return true
	// 	}
	// )
	// for name, assert := range ctx.req.Filter {
	// 	switch name {
	// 	// base [dir] of the path
	// 	case "dir":
	// 		{
	// 			s, _ := assert.(string)
	// 			whereDir(s)
	// 		}
	// 	// relative [path] match
	// 	case "path":
	// 		{
	// 			s, _ := assert.(string)
	// 			if customFilterIsPresent(s) {
	// 				break // skip ; all records has [path] assigned
	// 			}
	// 			dir, name := path.Split(s)
	// 			if dir != "" {
	// 				whereDir(dir)
	// 			}
	// 			whereName(name)
	// 		}
	// 	// path filename
	// 	case "repo", "id":
	// 		{
	// 			s, _ := assert.(string)
	// 			whereName(s)
	// 		}
	// 	// title ; lang specific
	// 	case "name", "title":
	// 		{
	// 			s, _ := assert.(string)
	// 			whereTitle(s)
	// 		}
	// 	// [NOT] GLOBAL ?
	// 	case "readonly":
	// 		{
	// 			if is, _ := assert.(bool); is {
	// 				// force "global" types ONLY !
	// 				// ignore "pdc" param value !
	// 				view.name = "sys"
	// 				view.expr = rdc.String() + " ISNULL"
	// 			} else {
	// 				// force "custom" types ONLY !
	// 				// ctx.params["pdc"] = domain.pdc // MAYBE: 0 !
	// 				view.name = "my"
	// 				view.expr = rdc.String() + " = :pdc"
	// 			}
	// 		}
	// 	// [NOT] GLOBAL & extendable ?
	// 	case "extendable":
	// 	case "available":
	// 		// LEFT JOIN pg_catalog.class
	// 	}
	// }
	// params["pdc"] = &pdc
	// query = query.Where(view.expr)
	query = ctx.datasetOptions.Apply(query, params)

	// ------- PAGING --------
	if size := ctx.req.GetSize(); size > 0 {
		// OFFSET (page-1)*size -- omit same-sized previous page(s) from result
		if page := ctx.req.GetPage(); page > 1 {
			query = query.Offset((uint64)((page - 1) * (size)))
		}
		// LIMIT (size+1) -- to indicate whether there are more result entries
		query = query.Limit((uint64)(size + 1))
	}

	ctx.Query = query
	ctx.plan = plan

	return nil
}

func customDatasetFetchRows(ctx *datasetQ, rows pgx.Rows, into *custompb.DatasetList) error {

	// cols, err := rows.Columns()
	// if err != nil {
	// 	return err
	// }
	var err error
	cols := rows.FieldDescriptions()
	var (
		e, n = 0, len(cols)
		stat error // *model.Error
		plan = ctx.plan
	)
	e = n // no "err" column
	// for e = 0; e < n && cols[e] != columnStatus; e++ {
	// 	// lookup: status."err" column requested ?
	// }
	if e < n {
		// "err" column found; requested !
		// inject internal decoder for that ...
		plan = append(plan[0:e+1], plan[e:]...)
		plan[e] = func(*custompb.Dataset) sql.Scanner {
			// // reacting ON status.err foreach node fetching ...
			// return contactsFetchStatus(&status).(SQLScanner)
			return ScanFunc(func(any) error { return nil })
		}
	}

	var (
		req  = ctx.req
		row  *custompb.Dataset        // active record
		scan = make([]any, len(plan)) // columns bound

		heap  []custompb.Dataset  // memory page
		data  []*custompb.Dataset // output data
		page  = into.GetData()    // input page
		limit = req.GetSize()     // limit count
	)
	// if page := req.GetPage(); page > 1 {
	into.Page = int32(req.GetPage())
	// }

	if 0 < limit {
		data = make([]*custompb.Dataset, 0, limit)
	}

	if n := limit - len(page); 1 < n {
		heap = make([]custompb.Dataset, n) // mempage; tidy
	}

	// FETCH
	var r, c int // [r]ow, [c]olumn
	for rows.Next() {
		// LIMIT
		if 0 < limit && len(data) == limit {
			// mark: more(!) record(s) available
			into.Next = true
			if into.Page < 1 {
				into.Page = 1 // default
			}
			break // rows.Next()
		}
		// RECORD
		row = nil // NEW
		if r < len(page) {
			// [INTO] given page records
			// [NOTE] order matters !
			row = page[r]
		} else if len(heap) > 0 {
			row = &heap[0]
			heap = heap[1:]
		}
		// ALLOC
		if row == nil {
			row = new(custompb.Dataset)
		}
		// BIND RECORD
		// for col, bind := range plan {
		// 	eval[col] = bind(node)
		// }
		c = 0
		for _, bind := range plan {
			cs := bind(row)
			if cs != nil {
				scan[c] = cs
				c++
				continue
			}
			// (df == nil)
			// omit; pseudo calc
		}
		// DECODE
		err = rows.Scan(scan...)
		if err != nil {
			break // rows.Next()
		}

		// (status.err NOTNULL) ?
		if stat != nil {
			err = stat
			break
		}
		// RESULT
		data = append(data, row)
		r++ // advance
	}

	if err == nil {
		err = rows.Err()
	}

	if err != nil {
		return err
	}

	if !into.Next && into.Page <= 1 {
		// The first page with NO more results !
		into.Page = 0 // Hide: NO paging !
	}
	into.Data = data
	return nil
}

func customDatasetFieldQuery(ctx *datasetQ, from, alias string) error {

	joinQ, err := customFieldSelectQuery(from) // fields: "+"
	if err != nil {
		return err
	}
	// Assert(MUST)
	cte := joinQ.Query.(sq.SelectBuilder)

	const (
		left  = aliasType
		right = aliasField
	)

	cte = cte.Where(
		fmt.Sprintf("%s = %s",
			sqlident{left, columnTypeId},
			sqlident{right, columnFieldOf},
		),
	)

	// for param, value := range joinQ.Params {
	// 	ctx.Params.set(param, value)
	// }

	query, _, err := cte.ToSql()
	if err != nil {
		return err
	}

	join := &JOIN{
		Kind: "LEFT JOIN",
		Source: "LATERAL(" +
			"SELECT ARRAY(" +
			// source +
			strings.Replace(
				strings.Replace(
					query,
					"SELECT ", "SELECT ROW(", 1,
				), " FROM", ") FROM", 1,
			) +
			"))",
		Alias: alias + "(list)",
		Pred:  "true",
	}
	// ctx.joins[alias] = join
	selectQ := ctx.Query.(sq.SelectBuilder)
	selectQ = selectQ.JoinClause(join)
	selectQ = selectQ.Column(sqlident{alias, "list"}.String())

	ctx.Query = selectQ
	ctx.plan = append(ctx.plan, func(row *custompb.Dataset) sql.Scanner {
		return ScanFunc(func(src any) error {
			var list List[custompb.Field]
			err := customScanFlatRows(&list, joinQ.plan, src)
			if err != nil {
				return err
			}
			row.Fields = append(row.Fields[:0], list.Data...)
			return nil
		})
	})
	return nil
}

func customFieldSelectQuery(from string, fields ...string) (cte *query[*custompb.Field], err error) {
	table := sqlident{from}
	if from == "" {
		table = sqlident{schemaCustom, tableField}
	}

	cte = newQuery[*custompb.Field]()
	cte.req.Fields = fields

	const (
		left = aliasField
	)
	var (
		query = psql.Select().
			From(fmt.Sprintf(
				"%s AS %s", table, left,
			)).
			// [FIELDS]: Mandatory(!); {id}
			Columns(
				sqlident{left, columnFieldName}.String(),
			)

		text pgtype.Text
		plan = dataScanPlan[*custompb.Field]{
			// [field].Id
			func(row *custompb.Field) sql.Scanner {
				return ScanFunc(func(src any) (err error) {
					row.Name = ""
					if src == nil {
						return nil
					}
					err = text.Scan(src)
					if err != nil {
						return // err
					}
					row.Id = text.String
					return nil
				})
			},
		}
	)

	if len(fields) == 0 {
		fields = []string{
			"id",   // "name"
			"name", // "title"
			"hint", // "usage"
			"type",
			"default",
			"readonly",
			"required",
			"disabled",
			"hidden",
		}
	}

	var cols = names{}
	for _, name := range fields {
		switch name {
		case "id":
			// core ; always !
		case "name":
			{
				const column = columnFieldTitle
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanText(&row.Name, src)
					})
				})
			}
		case "hint":
			{
				const column = columnFieldUsage
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanText(&row.Hint, src)
					})
				})
			}
		case "type", "kind":
			{
				const column = columnFieldTypeSpec
				if !cols.append(column) {
					continue // duplicate
				}
				// ------------------------------------------------ //
				// SELECT
				// 	fd.kind,
				// 	-- fd.type,
				// 	(
				// 		case
				// 		when fd.kind = 'list'
				// 		then jsonb_build_object
				// 		(
				// 			-- elem type descriptor --
				// 			fd.list, rtyp.spec
				// 		)
				// 		-- data type descriptor --
				// 		else rtyp.spec
				// 		end
				// 	) "type",
				// FROM custom.field fd
				// left join custom.dataset rel on fd.rel = rel.id
				// left join lateral
				// (
				// 	select
				// 	(
				// 		case
				// 		when rel.id notnull
				// 		then fd.type||jsonb_build_object
				// 		(
				// 			'name', rel.title,
				// 			'path', COALESCE((rel.dir||'/'),'')||rel.scope,
				// 			'primary', rel.primary,
				// 			'display', rel.display
				// 		)
				// 		else fd.type -- origin spec. --
				// 		end
				// 	) spec -- descriptor --
				// )
				// rtyp on true

				// resolved 'lookup' type kind reference !
				const right = "rel" // alias
				query = query.JoinClause(fmt.Sprintf(
					"LEFT JOIN %s %s ON %[2]s.%s = %s.%s",
					sqlident{schemaCustom, tableType}.String(),
					right, columnTypeId,
					left, columnFieldTypeLookup,
				))
				// actualize 'lookup' type reference descriptor !
				query = query.JoinClause(CompactSQL(fmt.Sprintf(
					`LEFT JOIN LATERAL
					(
						SELECT
						(
							CASE
							WHEN %[2]s.%[4]s NOTNULL
							THEN %[1]s.%[3]s || jsonb_build_object
							(
								'name', %[2]s.%[5]s,
								'path', COALESCE((%[2]s.%[6]s||'/'),'')||%[2]s.%[7]s,
								'primary', %[2]s.%[8]s,
								'display', %[2]s.%[9]s
							)
							ELSE %[1]s.%[3]s -- origin spec. --
							END
						) spec -- descriptor --
					)
					rtyp ON true`,
					// arguments
					left, right,
					// LEFT columns
					columnFieldTypeSpec,
					// RIGHT columns
					columnTypeId, columnTypeTitle,
					columnTypeDir, columnTypeName,
					columnTypeFieldPrimary, columnTypeFieldDisplay,
				)))
				// ------------------------------------------------ //
				query = query.Columns(
					sqlident{left, columnFieldTypeKind}.String(),
					// ident(left, "type"),
					CompactSQL(fmt.Sprintf(`(
						CASE
						WHEN %[1]s.%[2]s = 'list'
						THEN jsonb_build_object
						(
							-- elem type descriptor --
							%[1]s.%[3]s, rtyp.spec
						)
						-- data type descriptor --
						ELSE rtyp.spec
						END
					)`,
						// LEFT alias
						left,
						// LEFT column
						columnFieldTypeKind,
						columnFieldTypeList,
					)),
				)
				// [field].kind
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) (err error) {
						var name string
						err = customScanText(&name, src)
						if err != nil {
							return // err
						}
						code := datatypb.Kind_value[name]
						if code < 1 {
							return fmt.Errorf("cannot cast text %q into Kind", name)
						}
						row.Kind = datatypb.Kind(code)
						return nil
					})
				})
				// [field].type
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) (err error) {
						row.Type = nil
						if src == nil {
							return nil
						}
						mtyp := row.ProtoReflect()
						oneOf := mtyp.Descriptor().Fields().ByName(
							protoreflect.Name(row.Kind.String()),
						)
						// // if oneOf == nil {
						// // 	// "list" kind ?
						// // 	oneOf = mtyp.Descriptor().Fields().ByName(
						// // 		protoreflect.Name("type"), // union
						// // 	)
						// // }
						// varOf := mtyp.NewField(oneOf)
						// var varOf protoreflect.Value
						if oneOf != nil {
							// oneof [kind] field found !
							varOf := mtyp.NewField(oneOf)
							valOf := varOf.Message().Interface()
							err = customScanProtojsonPlan(&valOf).Scan(src)
							// defer func() {
							if err == nil {
								mtyp.Set(oneOf, varOf)
							}
							// }()
						} else if row.Kind == datatypb.Kind_list {
							// "list" [kind] special case !
							// Unmarshal into Field{oneof:type} structure !
							var listOf *custompb.Field
							err = customScanProtojsonPlan(&listOf).Scan(src)
							if err == nil {
								row.Type = listOf.GetType()
							}
						}
						return // err
						// // FieldDescriptor.Message()
						// // because type of each oneof field is the message type !
						// valOf := varOf.Message().Interface()
						// err = dbx.ScanProtoJSON(&valOf)(src)
						// if err != nil {
						// 	return err
						// }
						// mtyp.Set(oneOf, varOf)
						// return nil
					})
				})
			}
		case "default", "always":
			{
				const column = columnFieldDataDefault
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Columns(
					sqlident{left, columnFieldDataAlways}.String(),
					sqlident{left, columnFieldDataDefault}.String(),
				)
				// [field].always
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) (err error) {
						var is bool
						err = customScanBool(&is, src)
						if err != nil {
							return // err
						}
						if is {
							row.Value = &custompb.Field_Always{}
						} // else {
						// 	row.Value = &custompb.Field_Default{}
						// }
						return // nil
					})
				})
				// [field].default
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) (err error) {
						if src == nil {
							return nil
						}
						var (
							oneof = row.GetValue()
							vdata **structpb.Value
						)
						if set, is := oneof.(*custompb.Field_Always); !is {
							set := &custompb.Field_Default{}
							vdata = &set.Default
							oneof = set
						} else {
							vdata = &set.Always
						}
						// err := dbx.ScanProtoJSON(vdata)(src)
						// text := make([]byte, len(src)+2)
						// text[0] = '"'
						// copy(text[1:], src)
						// text[len(src)+2-1] = '"'
						err = customScanProtojsonPlan(vdata).Scan(src) // (text)
						if err != nil {
							row.Value = nil
							return // err
						}
						row.Value = oneof
						return nil
					})
				})
			}
		case "readonly":
			{
				const column = columnFieldIsReadonly
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanBool(&row.Readonly, src)
					})
				})
			}
		case "required":
			{
				const column = columnFieldIsRequired
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanBool(&row.Required, src)
					})
				})
			}
		case "disabled":
			{
				const column = columnFieldIsDisabled
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanBool(&row.Disabled, src)
					})
				})
			}
		case "hidden":
			{
				const column = columnFieldIsHidden
				if !cols.append(column) {
					continue // duplicate
				}
				query = query.Column(
					sqlident{left, column}.String(),
				)
				plan = append(plan, func(row *custompb.Field) sql.Scanner {
					return ScanFunc(func(src any) error {
						return customScanBool(&row.Hidden, src)
					})
				})
			}
		default:
			{
				// unknown field name
				err := custom.RequestError(
					"custom.dataset.field.invalid",
					"custom: dataset{fields{%s}} no such field",
					name,
				)
				return nil, err
			}
		}
	}

	query = query.OrderBy(
		sqlident{left, columnFieldOf}.String(),  // ASC
		sqlident{left, columnFieldNum}.String(), // ASC
	)

	cte.Query = query
	cte.plan = plan
	return cte, nil
}

// ------------------------------------------------------ //
//                       FILTER(s)                        //
// ------------------------------------------------------ //

func customFilterIsPresent(v string) bool {
	switch v {
	case "*", "":
		return true
	}
	return false
}

func customFilterIsSubstring(v string) bool {
	return 0 <= strings.IndexAny(v, "*?")
}

// [I]LIKE assertion value
func customFilterSubstringAssertion(v string) string {
	if !strings.ContainsAny(v, "_?%*") {
		return v // nothing todo
	}
	const ESC = '\\' // https://postgrespro.ru/docs/postgresql/12/functions-matching#FUNCTIONS-LIKE
	s := []byte(v)
	s = bytes.ReplaceAll(s, []byte{'_'}, []byte{ESC, '_'}) // escape control '_' (single char entry)
	s = bytes.ReplaceAll(s, []byte{'?'}, []byte{'_'})      // propagate '?' char for PostgreSQL purpose
	s = bytes.ReplaceAll(s, []byte{'%'}, []byte{ESC, '%'}) // escape control '%' (any char(s) or none)
	s = bytes.ReplaceAll(s, []byte{'*'}, []byte{'%'})      // propagate '*' char for PostgreSQL purpose
	return string(s)
}
