package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"

	// "github.com/micro/micro/v3/service/errors"
	custom "github.com/webitel/custom/data"
	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type RecordExtension interface {
	// GetCustom [extension] data record.
	GetCustom() *structpb.Struct
}

type RecordExtendable interface {
	RecordExtension
	// SetCustom [extension] data record.
	SetCustom(*structpb.Struct)
}

type extendableProto struct {
	record protoreflect.Message
	custom protoreflect.FieldDescriptor
}

// GetCustom [extension] data record.
func (m extendableProto) GetCustom() *structpb.Struct {
	return m.record.Get(m.custom).Message().Interface().(*structpb.Struct)
}

// SetCustom [extension] data record.
func (m extendableProto) SetCustom(data *structpb.Struct) {
	if data == nil {
		m.record.Clear(m.custom)
		return // NULL
	}
	m.record.Set(m.custom,
		protoreflect.ValueOf(
			data.ProtoReflect(),
		),
	)
}

// ProtoExtendable interface to deal with
// proto.(Message).Custom.(google.protobuf.Struct) field value
func ProtoExtendable(rec proto.Message) RecordExtendable {
	rmsg := rec.ProtoReflect()
	rtyp := rmsg.Descriptor()
	fd := rtyp.Fields().ByName("custom")
	if fd == nil {
		panic(fmt.Errorf("custom: record.(%s) is not extendable", rtyp.FullName()))
	}
	ok := false
	switch fd.Kind() {
	case protoreflect.MessageKind:
		{
			ok = (fd.Message().FullName() == (*structpb.Struct)(nil).ProtoReflect().Descriptor().FullName())
		}
	}
	if !ok {
		panic(fmt.Errorf("custom: record.(%s) is not extendable", rtyp.FullName()))
	}
	return extendableProto{
		record: rmsg,
		custom: fd,
	}
}

// ExtensionQueryBuilder for base [Dictionary] dataset query injections.
type ExtensionQueryBuilder interface {
	// Table relation name to the dataset records.
	Table() string
	// [from/left]    ; SELECT .. FROM contacts AS [left]
	// [table/right]  ; LEFT JOIN custom.x1_contacts AS [right] ON (left.id, left.dc) = (right.id, right.dc)
	Join(from SelectQ, left, table, right string) (query SelectQ, rel string)
	// [from]    ; SELECT .. FROM contacts AS [left]
	// [rel]     ; LEFT JOIN custom.x1_contacts AS [rel] ON (left.id, left.dc) = (right.id, right.dc)
	// [fields]  ; SELECT ROW([rel].fields,..)
	Columns(from SelectQ, rel string, fields ...string) (query SelectQ, scan func(RecordExtendable) sql.Scanner, err error)

	// // Where extension AND filter(s) ...
	// Where(from sq.SelectBuilder, rel string, filters map[string]any) (query sq.SelectBuilder, err error)
	// // OrderBy extension fields ...
	// OrderBy(from sq.SelectBuilder, rel string, sort ...string) (query sq.SelectBuilder, err error)

	// [pkx]       ; [P]rimary [K]ey [V]alue ; Accept: [SQLizer] -OR- GoValue
	// [data]      ; record changes to be saved !
	// [partial]   ; if [true] - updates given [data].field(s) only, otherwise - all known fields !
	Update(pkx any, data *structpb.Struct, partial bool) (query sq.Sqlizer, params Parameters, err error)
}

func (c *Catalog) Extension(as customrel.ExtensionDescriptor) (ExtensionQueryBuilder, error) {
	return newDataset(c, as)
}

// Dataset represents an interface for
// custom/reflect.DatasetDescriptor table
type dataset struct {
	dc    *Catalog
	rtyp  customrel.DatasetDescriptor
	table customTable // sqlident
}

func newDataset(dc *Catalog, rtyp customrel.DatasetDescriptor) (*dataset, error) {
	if rtyp == nil {
		panic("postgres.Dataset type descriptor is missing")
	}
	if rtyp.Dc() < 1 {
		panic("postgres.Dataset type descriptor is readonly")
	}
	ds := &dataset{
		dc: dc, rtyp: rtyp,
		table: customDatasetTable(rtyp),
		// 	rtyp.Dc(), rtyp.Name(),
		// ),
	}
	//
	// SELECT true
	// FROM information_schema.tables
	// WHERE table_schema = 'custom'
	//   AND table_name = 'd1_contacts'
	// ;
	//
	return ds, nil
}

var _ ExtensionQueryBuilder = (*dataset)(nil)

// Table relation name to the dataset records.
func (ds *dataset) Table() string {
	return ds.table.rel.String()
}

// [from/left]    ; SELECT .. FROM contacts AS [left]
// [table/right]  ; LEFT JOIN custom.x1_contacts AS [right] ON (left.id, left.dc) = (right.id, right.dc)
func (ds *dataset) Join(from SelectQ, left string, table string, right string) (query SelectQ, rel string) {
	// Extension
	// ext := ds.rtyp.(customrel.ExtensionDescriptor)
	// base := ext.Dictionary()
	// General
	// base := rtyp
	// default::TABLE
	relx := ds.table.rel
	if table != "" {
		// custom::CTE
		relx = sqlident(strings.Split(table, "."))
	}

	if right == "" {
		right = "x" // alias
	}

	// params := map[string]any{
	// 	"xdc": ext.Dc(),
	// }
	return from.JoinClause(fmt.Sprintf(
		"LEFT JOIN %[1]s AS %[2]s ON %[3]s.%[4]s = %[2]s.%[4]s AND %[3]s.%[5]s = %[2]s.%[5]s",
		relx.String(), right, left, ds.rtyp.Primary().Name(), columnDc,
	)), right
}

// [from]    ; SELECT .. FROM contacts AS [left]
// [rel]     ; LEFT JOIN custom.x1_contacts AS [rel] ON (left.id, left.dc) = (right.id, right.dc)
// [fields]  ; SELECT ROW([rel].fields,..)
func (ds *dataset) Columns(from SelectQ, rel string, fieldsQ ...string) (query SelectQ, scan func(RecordExtendable) sql.Scanner, err error) {
	// from, right = x.Join(from, left, right)

	// ext := ds.rtyp.(customrel.ExtensionDescriptor)
	// base := ext.Dictionary()
	// fieldsX := ext.Fields()
	// primary := base.Primary()

	var (
		// shorthands
		rtyp    = ds.rtyp
		fields  = rtyp.Fields()
		primary = rtyp.Primary()
		display = rtyp.Display()
		// is extension ?
		_, extensionIs = rtyp.(customrel.ExtensionDescriptor)
		// extpk    = false //
	)
	// PREPARE columns for SELECT !
	var (
		colhash  names
		columns  []string                     // = fields
		scanPlan dataScanPlan[*custom.Record] // fields
		addField = func(fd customrel.FieldDescriptor) error {
			// Do NOT select PRIMARY KEY value into result !
			// [NOTE]: is hidden from view ; available from [SUPER] data record !
			if extensionIs && (fd == primary || fd == display) { // fd.Name() == primary.Name() {
				// return nil // skip
				return fmt.Errorf("custom: extension.fields( %s ) no such field", fd.Name()) // hidden from data view !
			}
			if !colhash.append(fd.Name()) {
				return nil // duplicate !
			}
			vtyp := fd.Type()
			switch vtyp.Kind() {
			// case custom.LIST:
			case customrel.LIST:
				{
					elem := vtyp.(*custom.List).Elem()
					if _, ref := elem.(*custom.Lookup); !ref {
						// SCALAR[]
						columns = append(
							columns, sqlident{rel, fd.Name()}.String(),
						)
						scanPlan = append(
							scanPlan, customRecordScanFieldArray(fd),
						)
						break
					}
					// REFERENCE[]
					ref := struct {
						typof customrel.DictionaryDescriptor
						table customTable // sqlident
						alias string
						colpk sqlident // string
						coldn sqlident //string
						query SelectQ
					}{
						typof: elem.(*custom.Lookup).Dictionary(),
						alias: "e", // fmt.Sprintf("x%d", fd.Num()),
						query: psql.Select(),
					}
					ref.table = customDatasetTable(ref.typof)
					ref.colpk = sqlident{ref.alias, ref.typof.Primary().Name()}
					ref.query = ref.query.From(fmt.Sprintf(
						"%s %s", ref.table.rel, ref.alias,
					)).JoinClause(fmt.Sprintf(
						"JOIN UNNEST(%[1]s.%[2]s) WITH ORDINALITY AS vs(id, n) "+
							"ON %[3]s = vs.id AND %[4]s = %[1]s.dc",
						rel, fd.Name(),
						ref.colpk,
						sqlident{ref.alias, ref.table.dc},
					))
					ref.query, ref.coldn = ref.table.dn(ref.query, ref.alias, nil)
					// unescape: "::::"
					for i, n := 0, len(ref.coldn); i < n; i++ {
						ref.coldn[i], _, _ = BindNamed(ref.coldn[i], nil)
					}
					ref.query = ref.query.Column(fmt.Sprintf(
						"ARRAY_AGG((%s,%s)ORDER BY vs.n ASC)",
						ref.colpk, ref.coldn,
					))
					query, _, _ := ref.query.Prefix("(").Suffix(")").ToSql()
					right := fmt.Sprintf("x%d", fd.Num())
					join := JOIN{
						Kind:   "LEFT JOIN LATERAL",
						Source: query,
						Alias:  (right + "(list)"),
						Pred:   "true",
					}
					from = from.JoinClause(&join)
					columns = append(
						columns, sqlident{right, "list"}.String(),
					)
					scanPlan = append(
						scanPlan, customRecordScanFieldArray(fd),
					)
				}
			// case custom.LOOKUP:
			case customrel.LOOKUP:
				{
					// JOIN
					// SELECT
					ref := struct {
						typof customrel.DictionaryDescriptor
						table customTable // sqlident
						alias string
						colpk sqlident // string
						coldn sqlident // string
					}{
						typof: vtyp.(*custom.Lookup).Dictionary(),
						alias: fmt.Sprintf("x%d", fd.Num()),
					}
					// ref.table = customDatasetTable(rtyp.Dc(), ref.typof.Name())
					ref.table = customDatasetTable(ref.typof)
					ref.colpk = sqlident{ref.alias, ref.typof.Primary().Name()}
					from = from.JoinClause(fmt.Sprintf(
						"LEFT JOIN %[1]s %[2]s ON %[3]s.%[4]s = %[5]s AND %[3]s.dc = %[2]s.%[6]s",
						ref.table.rel.String(), ref.alias, // [RIGHT] custom.d$dc_$repo AS x$num
						rel, fd.Name(), // [LEFT] x.$fd
						ref.colpk, // [RIGHT] $repo.$pk
						ref.table.dc,
					))
					from, ref.coldn = ref.table.dn(from, ref.alias, nil)
					// unescape: "::::"
					for i, n := 0, len(ref.coldn); i < n; i++ {
						ref.coldn[i], _, _ = BindNamed(ref.coldn[i], nil)
					}
					// ROW(x_.%field)
					columns = append(columns, fmt.Sprintf(
						"(SELECT(%s,%s)WHERE %[1]s NOTNULL)",
						ref.colpk, ref.coldn,
					))
					scanPlan = append(scanPlan,
						dataScanFunc[*custom.Record](func(row *custom.Record) sql.Scanner {
							return ScanFunc(func(src interface{}) error {
								var ref *custompb.Lookup
								err := customScanLookup(&ref, fd.IsRequired()).Scan(src)
								if err != nil {
									return err
								}
								if ref == nil {
									// Skip <NULL> value(s) !
									return nil
								}
								return row.Set(fd, ref)
							})
						}))
				}
			default:
				{
					// [SCALAR]
					// // columns = append(columns, fd.Name())
					// // scanPlan = append(scanPlan, customRecordFieldScanPlan(fd))
					// columns[col] = ident(rel, fd.Name())
					// scanPlan[col] = customRecordFieldScanPlan(fd)
					column := sqlident{rel, fd.Name()}
					// .well-known ?
					if fd == display {
						from, column = ds.table.dn(from, rel, nil)
						// unescape: "::::"
						for i, n := 0, len(column); i < n; i++ {
							column[i], _, _ = BindNamed(column[i], nil)
						}
					}
					columns = append(columns, column.String())
					scanPlan = append(scanPlan, customRecordScanFieldValue(fd))
				}
			}
			return nil
		}
	)
	if len(fieldsQ) == 0 {
		// gather ALL known fields
		n := fields.Num()
		colhash = make(names, n)
		columns = make([]string, 0, n)
		scanPlan = make(dataScanPlan[*custom.Record], 0, n)
		fields.Range(func(fd customrel.FieldDescriptor) bool {
			err = addField(fd)
			return err == nil
		})
		if err != nil {
			return // from, nil, err
		}
	} else {
		// validate fields input
		n := len(fieldsQ)
		colhash = make(names, n)
		columns = make([]string, 0, n)
		scanPlan = make(dataScanPlan[*custom.Record], 0, n)
		// [FIXME]: duplicate columns !!!
		var fd customrel.FieldDescriptor
		for _, name := range fieldsQ {
			if fd = fields.ByName(name); fd == nil {
				// Field Not Found !
				err = custom.RequestError(
					"custom.extension.field.not_found",
					"custom: extension( field: %s ); no such field",
					name,
				)
				return // from, nil, err
			}
			// scanPlan = append(scanPlan, customRecordFieldScanPlan(fd))
			err = addField(fd)
			if err != nil {
				return // from, nil, err
			}
		}
	}
	// if err != nil {
	// 	return // from, scan, err
	// }
	var (
		sep string
		row strings.Builder
	)
	defer row.Reset()
	// row.WriteByte('(')
	row.WriteString("(SELECT ROW(")
	for _, expr := range columns {
		row.WriteString(sep)
		row.WriteString(expr)
		sep = ","
	}
	// row.WriteByte(')')
	fmt.Fprintf(&row, ") WHERE %s.%s NOTNULL)", rel, primary.Name()) // AS "custom"
	from = from.Column(row.String())
	scan = func(rec RecordExtendable) sql.Scanner {
		return ScanFunc(func(src any) (err error) {
			if src == nil {
				return // NULL
			}
			row := custom.NewRecord(rtyp) // ext.NewRecord()
			err = customRecordScanFlatRow(scanPlan)(row).Scan(src)
			if err != nil {
				// CAST [sql] TO [custom] types failed !
				return err
			}
			// err = row.Err()
			// if err != nil {
			// 	// [custom] composite (extension) type values vilation !
			// 	return err
			// }
			data := row.Proto()
			rec.SetCustom(data)
			return // err
		})
	}
	return from, scan, nil
}

// [oid]       ; [P]rimary [K]ey [V]alue ; Accept: [SQLizer] -OR- GoValue
// [data]      ; record changes to be saved !
// [partial]   ; if [true] - updates given [data].field(s) only, otherwise - all known fields !
func (ds *dataset) Update(oid any, data *structpb.Struct, partial bool) (query Sqlizer, params Parameters, err error) {

	const (
		// patch = false
		paramDc = "xdc"
		paramPk = "xpk"
	)

	var (
		dataset = ds.rtyp           // x.Extension
		primary = dataset.Primary() // dataset.TypeOf().Primary()
	)

	params = map[string]any{
		paramDc: dataset.Dc(),
	}

	switch as := oid.(type) {
	case sq.SelectBuilder:
		{
			// ( SELECT id FROM updated LIMIT 1 )
			oid = as.Prefix("(").Limit(1).Suffix(")")
		}
	// case []any:
	default:
		{
			// If NOT [SQL] - make parameter ..
			if _, is := oid.(sq.Sqlizer); !is {
				params.Add(paramPk, oid) // (sql.Valuer) !
				oid = sq.Expr((":" + paramPk))
			}
		}
	}

	// Unmarshal [application/json+proto] record data !
	var record = custom.NewRecord(dataset)
	for name, value := range data.GetFields() {
		// Find field by name ..
		fd := dataset.Fields().ByName(name)
		if fd == nil {
			// No such field !
			err = custom.RequestError(
				"custom.extensions.field.not_found",
				"custom: %s{%s} no such field",
				dataset.Path(), name,
			)
			return // err
		}
		// Accept field value spec. ?
		err = record.Set(fd, value)
		if err != nil {
			jsonv, _ := value.MarshalJSON()
			err = custom.RequestError(
				"custom.extensions.field.bad_value",
				"custom: %s[1].custom(%s).value(%s) ; error: %v",
				dataset.Path(), name, jsonv, err,
			)
			return // err
		}
	}
	// Has at least SOME field value changes ?
	if partial && len(record.Fields()) == 0 {
		// No changes to perform !
		return nil, nil, nil
	}
	// [TODO] WITH DEFAULTS ...
	// [NOTE] MAY populate additional field(s) values

	// PREPARE SQL
	const (
		rel = "e" // alias
	)
	var (
		columns []string
		fields  []customrel.FieldDescriptor
		values  []any

		updateQ strings.Builder
		// relTable = customDatasetTable(dataset.Dc(), dataset.Name())
		relTable = customDatasetTable(dataset)
	)
	defer updateQ.Reset() // Dispose()
	_, _ = fmt.Fprintf(&updateQ,
		"ON CONFLICT (%s) DO UPDATE SET"+
			" ver = (%[2]s.ver + 1)", // [FIXME]: + (OLD.* IS DISTINCT FROM NEW.*) ? 1 : 0
		primary.Name(), rel,
	)
	n := len(record.Fields())
	if !partial {
		n = dataset.Fields().Num()
	}
	columns = make([]string, 0, n+2)
	fields = make([]customrel.FieldDescriptor, 0, n+2)
	values = make([]any, 0, n+2)
	// [INSERT] ( DC, PK )
	columns = append(columns, "dc")
	values = append(values, sq.Expr((":" + paramDc)))
	columns = append(columns, strconv.Quote(primary.Name()))
	values = append(values, oid)
	recordValue := func(fd customrel.FieldDescriptor, vs any) error {
		// Cast to (sql.Value) ..
		vs, err = CustomTypeSqlValue(fd.Type(), vs)
		if err != nil {
			return err
		}
		// INSERT INTO
		columns = append(columns, strconv.Quote(fd.Name()))
		fields = append(fields, fd)
		// VALUE(S)
		param := "r1c" + strconv.Itoa(fd.Num()) // fd.Name()
		params[param] = vs
		values = append(values, sq.Expr((":" + param)))
		// ON CONFLICT DO UPDATE SET
		_, _ = fmt.Fprintf(&updateQ,
			", %[1]q = EXCLUDED.%[1]s",
			fd.Name(),
		)
		return nil
	}
	// PROJECT: fields | values
	if partial {
		// Walk thru populated fields ONLY !
		record.Range(func(fd customrel.FieldDescriptor, vs any) bool {
			err = recordValue(fd, vs)
			return err == nil
		})
	} else {
		// Walk thru ALL known fields ...
		dataset.Fields().Range(func(fd customrel.FieldDescriptor) bool {
			// omit [P]rimary [K]ey changes !
			if fd.Name() == primary.Name() {
				return true
			}
			// Get record value
			err = recordValue(fd, record.Get(fd))
			return err == nil
		})

	}
	// Cast field value failed !
	if err != nil {
		return // query, params, err
	}

	// INSERT INTO ...
	query = psql.
		Insert(fmt.Sprintf(
			"%s AS %s",
			relTable.rel.String(), rel,
		)).
		Columns(columns...).
		Values(values...).
		// Suffix(fmt.Sprintf(
		// 	"WHERE %[1]s.%s = :x_pk AND %[1]s.dc = :x_dc",
		// 	rel, PK.Name(),
		// )).
		Suffix(
			// ON CONFLICT (PK) DO UPDATE SET ...
			updateQ.String(),
		).
		Suffix(fmt.Sprintf(
			"RETURNING %s.*",
			rel,
		))

	return query, params, nil
}

func customRecordScanFieldValue(fd customrel.FieldDescriptor) dataScanFunc[*custom.Record] {
	return dataScanFunc[*custom.Record](func(row *custom.Record) sql.Scanner {
		return ScanFunc(func(src any) error {
			if src == nil {
				// NULL !
				return nil
			}
			// as := fd.Type()
			// rv := as.New()
			// err := rv.Decode(src)
			// if err == nil {
			// 	err = row.Set(fd, rv.Interface())
			// }
			// return err
			return row.Set(fd, src)
		})
	})
}

func customRecordScanFieldArray(fd customrel.FieldDescriptor) dataScanFunc[*custom.Record] {
	return dataScanFunc[*custom.Record](func(row *custom.Record) sql.Scanner {
		return ScanFunc(func(src any) (err error) {

			if src == nil {
				return // nil
			}

			var (
				rows   pgtype.Array[pgtype.UndecodedBytes]
				input  []byte
				format int16 = pgtype.BinaryFormatCode
				mtypes       = pgtype.NewMap()
			)
			switch data := src.(type) {
			case []byte:
				{
					if len(data) == 0 {
						return // NULL
					}
					input = data
					format = pgtype.BinaryFormatCode
				}
			case string:
				{
					if len(data) == 0 || data == "{}" {
						return // NULL
					}
					input = []byte(data)
					format = pgtype.TextFormatCode
				}
			default:
				// return errors.InternalServerError(
				// 	"store.postgres.fetch.data.convert",
				// 	"postgres: convert %[1]T into []*custom.Field",
				// 	src,
				// )
				return fmt.Errorf("postgres: convert %[1]T into []%s", src, fd.Type().(*custom.List).Elem().Kind())
			}
			// parse: array[record]
			err = mtypes.PlanScan(
				pgtype.OIDArrayOID,
				// pgtype5.RecordArrayOID,
				format, &rows,
			).Scan(
				input, &rows,
			)
			if err != nil {
				return // err
			}

			var (
				size = len(rows.Elements)
			)

			if size == 0 {
				return row.Set(fd, nil)
			}

			elem := fd.Type().(*custom.List).Elem()
			list := reflect.MakeSlice(
				reflect.SliceOf(reflect.TypeOf(
					elem.New().Interface(),
				)),
				0, size,
			)
			// fmt.Printf("ARRAY%T\n", list.Interface())

			var plan sql.Scanner
			if _, is := elem.(*custom.Lookup); is {
				plan = ScanFunc(func(src any) (err error) {
					// [NOTE]: SELECT returns VALID reference records ONLY !
					const notnull = true
					var item *custompb.Lookup
					err = customScanLookup(
						&item, notnull, "id", "name",
					).Scan(src)
					if err == nil {
						list = reflect.Append(
							list, reflect.ValueOf(
								item,
							),
						)
					}
					return // err?
				})
			} else {
				rtyp, _ := customDataType(elem)
				plan = ScanFunc(func(src any) (err error) {
					// item := elemT.New()
					// mtypes.PlanScan(0, format, )
					switch format {
					case pgtype.TextFormatCode:
						src = []byte(src.(string))
					case pgtype.BinaryFormatCode:
						src = src.([]byte)
					}
					src, err = rtyp.Codec.DecodeValue(
						pgtypes, rtyp.OID, format, src.([]byte), // [FIXME]
					)
					if err != nil {
						return // err
					}
					item := elem.New()
					err = item.Decode(src)
					if err != nil {
						// Failed to decode PostgreSQL ARRAY element !
						return // err
					}
					list = reflect.Append(
						list, reflect.ValueOf(
							item.Interface(),
						),
					)
					return // err
				})
			}

			for _, elem := range rows.Elements {
				// pgtype5.BinaryFormatCode
				src := any([]byte(elem)) // binary
				if format == pgtype.TextFormatCode {
					src = string(elem) // text
				}
				err = plan.Scan(src)
				if err != nil {
					// Failed to decode PostgreSQL ARRAY element !
					return // err
				}
			}

			if err == nil {
				err = row.Set(fd, list.Interface())
			}
			return err
		})
	})
}

func customRecordScanFlatRow(plan dataScanPlan[*custom.Record]) dataScanFunc[*custom.Record] {
	return dataScanFunc[*custom.Record](func(row *custom.Record) sql.Scanner {
		return ScanFunc(func(src interface{}) (err error) {
			if src == nil {
				return // nil
			}
			var (
				// input  []byte
				format int16 // = pgtype5.TextFormatCode
				binary datatypeScanner
				record compositeScanner
			)
			switch data := src.(type) {
			case string:
				{
					switch data {
					case "", "()":
						return // NULL
					}
					// input = []byte(data)
					format = pgtype.TextFormatCode
					record = pgtype.NewCompositeTextScanner(
						nil, []byte(data),
					)
				}
			case []byte:
				{
					if len(data) == 0 {
						return // NULL
					}
					// input = data
					format = pgtype.BinaryFormatCode
					record = pgtype.NewCompositeBinaryScanner(
						nil, []byte(data),
					)
					binary = record.(datatypeScanner)
				}
			default:
				{
					// err = errors.InternalServerError(
					// 	"store.postgres.fetch.data.convert",
					// 	"postgres: could not convert %[1]T into *Record",
					// 	src,
					// )
					err = fmt.Errorf("postgres: convert %[1]T into *custom.Record", src)
					return // err
				}
			}
			var (
				oid uint32 = pgtype.TextOID // [FIXME] !!!
				typ *pgtype.Type
				val any
			)
			for _, col := range plan {
				if !record.Next() {
					err = record.Err()
					if err != nil {
						return // err
					}
					// FIXME: no more data available ?
					break
				}
				if binary != nil {
					oid = binary.OID()
				}
				typ, _ = pgtypes.TypeForOID(oid)
				val, err = typ.Codec.DecodeDatabaseSQLValue(
					pgtypes, oid, format, record.Bytes(),
				)
				if err != nil {
					return // err
				}
				err = col(row).Scan(val)
				if err != nil {
					return // err
				}
				// err = dtype.Codec.PlanScan(
				// 	pgx5type, dtoid, format, value,
				// ).Scan(
				// 	record.Bytes(), value,
				// )
				// if err != nil {
				// 	return // err
				// }
			}
			err = record.Err()
			return // err
		})
	})
}
