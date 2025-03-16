package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	// proto1 "github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	custom "github.com/webitel/custom/data"
	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var pgtypes = pgtype.NewMap()

type compositeScanner interface {
	Err() error
	Next() bool
	Bytes() []byte
}

type datatypeScanner interface {
	OID() uint32
}

func customScanBool(dst *bool, src any) (err error) {
	*dst = false
	if src == nil {
		return // nil
	}
	var data pgtype.Bool
	err = data.Scan(src)
	if err != nil {
		return // err
	}
	if data.Valid {
		*dst = data.Bool
	}
	return // nil
}

// func customScanText(dst *string, src any) (err error) {
// 	*dst = ""
// 	if src == nil {
// 		return // nil
// 	}
// 	var data pgtype.Text
// 	err = data.Scan(src)
// 	if err != nil {
// 		return // err
// 	}
// 	if data.Valid {
// 		*dst = data.String
// 	}
// 	return // nil
// 	// switch e := src.(type) {
// 	// case string:
// 	// 	*dst = e
// 	// case []byte:
// 	// 	if len(e) > 0 {
// 	// 		*dst = string(e)
// 	// 	}
// 	// case int64:
// 	// 	if e != 0 {
// 	// 		*dst = strconv.FormatInt(e, 10)
// 	// 	}
// 	// default:
// 	// 	err = fmt.Errorf(
// 	// 		"postgres: convert %T value into string", src,
// 	// 	)
// 	// }
// 	// return // err
// }

func customScanText(dst *string, src any) (err error) {
	return (*zeronull.Text)(dst).Scan(src)
}

func customScanAsText(dst *string, src any) (err error) {
	*dst = ""
	if src == nil {
		return // nil
	}
	// data, ok := pgtypes.TypeForValue(src)
	// data.Codec.DecodeValue(pgtypes, )
	switch e := src.(type) {
	case string:
		*dst = e
	case []byte:
		if len(e) > 0 {
			*dst = string(e)
		}
	case int64:
		if e != 0 {
			*dst = strconv.FormatInt(e, 10)
		}
	default:
		err = fmt.Errorf("cannot scan %T value into string", src)
	}
	return // err
}

func customScanLookup(into **custompb.Lookup, notnull bool, fields ...string) sql.Scanner {
	ref := (*into) // cached
	deref := func() *custompb.Lookup {
		if ref == nil {
			ref = new(custompb.Lookup)
		}
		return ref
	}
	defer func() {
		if notnull {
			if deref().Name == "" {
				ref.Name = "[deleted]"
			}
			(*into) = ref
		}
	}()
	return ScanFunc(func(src interface{}) (err error) {
		defer func() {
			if err != nil {
				ref = nil
			}
			if ref != nil &&
				ref.Id == "" &&
				ref.Name == "" &&
				ref.Type == "" {
				// empty ; zero
				ref = nil
			}
			// assign value !
			(*into) = ref
		}()
		if src == nil {
			ref = nil
			return // nil
		}
		var (
			// input  []byte
			format int16 = pgtype.TextFormatCode
			binary datatypeScanner
			record compositeScanner
		)
		switch data := src.(type) {
		case string:
			{
				switch data {
				case "", "()":
					ref = nil
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
					ref = nil
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
				err = fmt.Errorf("cannot scan %T into *Lookup", src)
				return // err
			}
		}
		if len(fields) == 0 {
			fields = []string{
				"id", "name",
			}
		}
		var (
			oid uint32 = pgtype.TextOID // default
			typ *pgtype.Type
			val any
		)
		for _, fd := range fields {
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
			typ, _ = pgtypes.TypeForOID(oid) // MUST
			val, err = typ.Codec.DecodeDatabaseSQLValue(
				pgtypes, oid, format, record.Bytes(),
			)
			if err != nil {
				return // err
			}
			switch fd {
			case "id":
				// pgText{&deref().Id}.Scan(val)
				err = customScanAsText(&deref().Id, val)
			case "name":
				// pgText{&deref().Name}.Scan(val)
				err = customScanText(&deref().Name, val)
			case "type":
				// pgText{&deref().Type}.Scan(val)
				err = customScanText(&deref().Type, val)
			default:
				err = fmt.Errorf("cannot scan %T into Lookup{%s}; no such field", val, fd)
				// return err
			}
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
		return // nil
	})
}

var (
	// proto1MessageType = reflect.TypeOf((*proto1.Message)(nil)).Elem()
	protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

func customScanProtojsonPlan(dst any) sql.Scanner {
	return ScanFunc(func(src any) (err error) {
		if src == nil {
			return nil
		}
		var (
			// err  error
			data []byte
		)
		switch v := src.(type) {
		case json.RawMessage:
			data = v
		case []byte:
			data = v
		default:
			return fmt.Errorf("cannot cast %T into %T", src, dst)
		}

		if len(data) == 0 {
			return nil // NULL
		}

		rv := reflect.ValueOf(dst)
		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("cannot cast %T into %T; expect pointer to proto.Message", src, dst)
		}

		rv = reflect.Indirect(rv)
		var (
			rt = rv.Type()
			// mi func() interface{}
			mp proto.Message
			rp protoreflect.Message
		)

		if rt.Implements(protoMessageType) {
			mp = rv.Interface().(proto.Message)
			rp = mp.ProtoReflect()
			if !rp.IsValid() {
				rp = rp.New()
				rv.Set(reflect.ValueOf(
					rp.Interface(),
				))
			}
		} // } else if rt.Implements(proto1MessageType) {
		// 	mp = proto1.MessageV2(rv.Interface())
		// 	rp = mp.ProtoReflect()
		// 	if !rp.IsValid() {
		// 		rp = rp.New()
		// 		rv.Set(reflect.ValueOf(
		// 			proto1.MessageV1(rp.Interface()),
		// 		))
		// 	}

		if rp != nil {
			err = protojson.UnmarshalOptions{
				AllowPartial:   true,
				DiscardUnknown: true,
				Resolver:       nil,
			}.Unmarshal(data, rp.Interface())
		} else {
			err = json.Unmarshal(data, dst)
		}

		return err
	})
}

type List[TRow any] struct {
	Data []*TRow `json:"data,omitempty"`
	Page int     `json:"page,omitempty"`
	Next bool    `json:"next,omitempty"`
}

// [SCAN]: '{"ROW(<column(s),..>)","ROW(..)"}'::contact_label[]
func customScanFlatRows[TRow any](dst *List[TRow], plan dataScanPlan[*TRow], src any) (err error) {
	if src == nil {
		return // NULL
	}
	var (
		rows   pgtype.Array[pgtype.UndecodedBytes]
		input  []byte
		format int16 = pgtype.BinaryFormatCode
		// mtypes       = pgtype5.NewMap()
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
		return fmt.Errorf("cannot cast %T into []%T", src, reflect.TypeFor[TRow]().Name())
	}
	// parse: array[record]
	err = pgtypes.PlanScan(
		pgtype.RecordArrayOID,
		format,
		&rows,
	).Scan(
		input,
		&rows,
	)
	if err != nil {
		return // err
	}

	var (
		node *TRow      // *pbdata.Field  // row  *pb.Label
		heap []TRow     // mem  []pb.Label
		list = dst.Data // row.GetFields() // pb.LabelList // []*pb.Label
		page = list     // .GetData()  // input
		data []*TRow

		rtyp = reflect.TypeFor[TRow]()
		size = len(rows.Elements)
		// limit = int(req.Args.Size())
		limit = size

		rec interface {
			Err() error
			Next() bool
			Bytes() []byte
		}
	)

	if size == 0 {
		// if list != nil {
		dst.Data = nil
		dst.Next = false
		// 	pnum := int32(req.Page())
		// 	if pnum > 1 {
		// 		list.Page = pnum
		// 	}
		// }
		return nil
	}

	// if list == nil {
	// 	list = new(pbdata.FieldList)
	// }
	// // Output: page number
	// list.Page = int32(req.Page())

	if 0 < size {
		data = make([]*TRow, 0, size)
	}

	if n := limit - len(page); 1 < n {
		heap = make([]TRow, n) // mempage; tidy
	}

	// // DECODE
	// // var r, c int // [r]ow, [c]olumn
	// var statusErr *model.Error // err
	for r, bin := range rows.Elements {
		// LIMIT
		if 0 < limit && limit == len(data) {
			dst.Next = true
			if dst.Page < 1 {
				dst.Page = 1
			}
			// data = append(data, nil) // next: true
			break
		}
		// RECORD
		// // node = reflect.Indirect(reflect.New(reflect.TypeFor[TRow]())).Interface().(TRow)
		// node = reflect.New(rtyp).Interface().(*TRow)
		node = nil // NEW
		if r < len(page) {
			// [INTO] given page records
			// [NOTE] order matters !
			node = page[r]
		} else if len(heap) > 0 {
			node = &heap[0]
			heap = heap[1:]
		}
		// ALLOC
		if node == nil {
			node = reflect.New(rtyp).Interface().(*TRow)
		}
		// DECODE
		// raw := pgtype.NewCompositeTextScanner(nil, []byte(elem))
		switch format {
		case pgtype.TextFormatCode:
			rec = pgtype.NewCompositeTextScanner(pgtypes, []byte(bin))
		case pgtype.BinaryFormatCode:
			rec = pgtype.NewCompositeBinaryScanner(pgtypes, []byte(bin))
		default:
			rec = nil
		}
		// for c, bind := range plan {
		var oid uint32 = pgtype.TextOID // column value data type
		for _, bind := range plan {
			if !rec.Next() { /// .ScanValue calls .Next(!)
				// break
				return rec.Err()
			}

			binary, _ := rec.(interface {
				OID() uint32
			})
			if binary != nil {
				oid = binary.OID()
			}

			// // [STATUS] err ?
			// if c == 0 && status {
			// 	raw.ScanDecoder(
			// 		contactsFetchStatus(&statusErr),
			// 	)
			// 	if err == nil && statusErr != nil {
			// 		err = statusErr
			// 	}
			// 	if err != nil {
			// 		return
			// 	}
			// }

			df := bind(node)
			if df == nil {
				// omit; pseudo calc
				continue
			}
			// mtypes.PlanScan()
			// raw.ScanValue(df)
			typof, _ := pgtypes.TypeForOID(oid)
			value, re := typof.Codec.DecodeDatabaseSQLValue(
				pgtypes, oid, format, rec.Bytes(),
			)
			if err = re; err != nil {
				return // err
			}
			err = df.Scan(value)
			if err != nil {
				return // err
			}
		}
		data = append(data, node)
	}
	if !dst.Next && dst.Page <= 1 {
		// The first page with NO more results !
		dst.Page = 0 // Hide: NO paging !
	}
	dst.Data = data
	return // err
}

func customTypeSqlValue(dt customrel.Type, vs any) (v any, err error) {
	if vs == nil {
		// NULL
		return nil, nil
	}
	switch v := vs.(type) {
	case *custompb.Lookup:
		{
			if v.GetId() == "" {
				// NULL
				return nil, nil
			}
			// switch fd.Kind() {}
			rt := dt.(*custom.Lookup)       // reference.(lookup) type
			pk := rt.Dictionary().Primary() // reference.[primary] field
			rv := pk.Type().New()           // value type codec
			err := rv.Decode(v.GetId())
			if err != nil {
				// failed to cast string value to it's data type !
				return v.GetId(), err // string value !
			}
			// return rv.Interface(), nil // [primary] key type Go value !
			vs = rv.Interface() // NULL(-able) !
			// process as indirect below !
		}
	}
	// reflect
	rv := reflect.ValueOf(vs)
	if !rv.IsValid() {
		// [untyped]: NULL !
		return nil, nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			// [typical]: NULL !
			return nil, nil
		}
		// *int64
		// *uint64
		// *tim.Time
		rv = reflect.Indirect(rv)
	}
	switch rv.Kind() {
	case reflect.Slice:
		{
			listT := dt.(*custom.List)
			itemT := listT.Elem()

			sizeV := rv.Len()
			listV := make([]any, sizeV)

			for i := 0; i < sizeV; i++ {
				listV[i], err = customTypeSqlValue(
					itemT, rv.Index(i).Interface(),
				)
				if err != nil {
					// failed to cast list[item] value !
					return nil, err
				}
			}
			return pgtype.FlatArray[any](listV), nil
		}
	}
	// NULL(-able) primitive value !
	return rv.Interface(), nil
}

func customTypeName(typ customrel.Type) (name string) {
	switch typ.Kind() {
	case customrel.LOOKUP:
		{
			ref := typ.(*custom.Lookup)
			rel := ref.Dictionary()
			typ = rel.Primary().Type()
			return customTypeName(typ)
		}
	case customrel.LIST:
		{
			typ = typ.(*custom.List).Elem()
			name = "_" + customTypeName(typ) // array
		}
	case customrel.BOOL:
		name = "bool"
	case customrel.INT, customrel.INT32:
		name = "int4"
	case customrel.INT64:
		name = "int8"
	case customrel.UINT, customrel.UINT32:
		name = "int4"
	case customrel.UINT64:
		name = "int8"
	case customrel.FLOAT, customrel.FLOAT32:
		name = "float4"
	case customrel.FLOAT64:
		name = "float8"
	case customrel.BINARY:
		name = "bytea"
	case customrel.STRING:
		name = "text"
	case customrel.RICHTEXT:
		name = "text"
	case customrel.DATETIME:
		name = "timestamp" // [ without time zone ]
	case customrel.DURATION:
		name = "interval"
	// case customrel.NONE:
	default:
		name = "text" // [FIXME] !!!
	}
	return // name
}

func customDataType(typ customrel.Type) (*pgtype.Type, bool) {
	return pgtypes.TypeForName(customTypeName(typ))
}
