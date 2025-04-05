package data

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/types/known/structpb"
)

type Record struct {
	mu     sync.Mutex // protects mutations
	typeof customrel.DatasetDescriptor
	values []any    // map[string]any
	fields []string // changed(!)
}

func NewRecord(typeof customrel.DatasetDescriptor) *Record {
	return &Record{
		typeof: typeof,
		values: make([]any, typeof.Fields().Num()),
	}
}

var _ customrel.Record = (*Record)(nil)

func (e *Record) Custom(pragma.DoNotImplement) {}

// Dataset returns type of the data structure
func (e *Record) Dataset() customrel.DatasetDescriptor {
	return e.typeof
}

// Get Field's value
func (e *Record) Get(fd customrel.FieldDescriptor) (v any) {

	e.mu.Lock()
	defer e.mu.Unlock()

	off := fd.Num() - 1 // offset
	if e.typeof.Fields().Get(off) != fd {
		return fmt.Errorf("custom: record.get( field: %s ); invalid descriptor", fd.Name())
	}
	return e.values[off]
}

// keep LOCKED
func (e *Record) set(fd customrel.FieldDescriptor, val any) error {
	// panic("not implemented")

	// [UPDATE] field value
	off := (fd.Num() - 1)
	e.values[off] = val
	// remember field name been updated
	name := fd.Name()
	i, n := 0, len(e.fields)
	for ; i < n && e.fields[i] != name; i++ {
		// lookup: field duplicate !
	}
	// Not Found ?
	if i == n {
		// first time update !
		e.fields = append(e.fields, name)
		// [NOTE] ( i < n )
	} else if i < (n - 1) {
		// [MOVE] last ; keep sequence for history
		copy(e.fields[i:], e.fields[i+1:])
		e.fields[n-1] = name
	}
	return nil
}

func (e *Record) Set(fd customrel.FieldDescriptor, val any) error {

	e.mu.Lock()
	defer e.mu.Unlock()
	// [CHECK] known data field
	off := fd.Num() - 1 // offset
	if e.typeof.Fields().Get(off) != fd {
		return fmt.Errorf("custom: record.set( field: %s ); invalid descriptor", fd.Name())
	}
	// [TODO]: distinguish two different modes, like:
	// - [enduser].Set(value)
	// - [default].Set(value)
	// [DESIGN]: e.BeginUpdate(); e.EndUpdate(); methods ???
	// // CONSTRAINT: DISABLED
	// if fd.IsDisabled() { // is [user] set ? or [default] value ?
	// 	return RequestError(
	// 		"custom.field.disabled.violation",
	// 		"custom: field( %s ) is disabled",
	// 		fd.Name(),
	// 	)
	// }
	// [CHECK] data type constrains
	rv := fd.Type().New()
	// CONSTRAINT TYPE
	err := rv.Decode(val)
	if err != nil {
		// Field Type Value specification invalid
		return err
	}
	// CONSTRAINT VALUE
	err = rv.Err()
	if err != nil {
		// Field Value Type constraints violation
		return err
	}
	// normalized !
	val = rv.Interface()
	// // CONSTRAINT: REQUIRED
	// if fd.IsRequired() {
	// 	// must := fd.Default().Always()
	// 	if customrel.IsNull(val) { // [FIXME]: || IsZero(?)
	// 		return RequestError(
	// 			"custom.field.required.violation",
	// 			"custom: field(%s) value required but missing",
	// 			fd.Name(),
	// 		)
	// 	}
	// }
	// UPDATE ; SAVE
	return e.set(fd, val)
}

// func (e *Record) BeginUpdate() {
// 	e.ref.Lock()
// 	defer e.ref.Unlock()
// 	e.fields = nil
// }

// func (e *Record) EndUpdate() {
// 	// e.ref.Lock()
// 	// defer e.ref.Unlock()
// 	// e.fields = nil
// }

// A set of fields whose values ​​have been updated !
func (e *Record) Fields() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := len(e.fields)
	if n == 0 {
		return nil
	}
	set := make([]string, n)
	copy(set, e.fields)
	return set
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
// func (rec *Record) Range(walk func(fd Field, vs Value) bool) {
func (e *Record) Range(next func(fd customrel.FieldDescriptor, vs any) bool) {

	// e.mu.Lock()
	// defer e.mu.Unlock()

	fields := e.typeof.Fields()
	for _, name := range e.fields {
		fd := fields.ByName(name)
		if !next(fd, e.values[fd.Num()-1]) {
			break // return
		}
	}
	// for fd, v := range e.values {
	// 	if v == nil {
	// 		continue // not populated
	// 	}
	// 	if !next(fields.Get(fd), v) {
	// 		break // return
	// 	}
	// }
}

func mapValue(v any) any {
	if v == nil {
		return nil
	}
	switch e := v.(type) {
	case *custompb.Lookup:
		{
			if e == nil {
				return nil // untyped
			}
			object := map[string]any{
				"id":   e.Id,
				"name": e.Name,
				"type": e.Type,
			}
			isnull := true
			for fd, vs := range object {
				if vs == "" {
					delete(object, fd)
					continue
				}
				isnull = false
			}
			v = nil // NULL
			if !isnull {
				v = object // map[string]any
			}
			return v
		}
	case *time.Duration:
		{
			if e == nil {
				return nil // untyped
			}
			return CastDurationAsNumber(*e)
		}
	case *time.Time:
		{
			if e == nil {
				return nil // untyped
			}
			return CastDateTimeAsNumber(*e)
		}
	}
	// reflect Value.(Nullable) !
	rv := reflect.ValueOf(v)
	// ( val == nil ) ?
	if !rv.IsValid() {
		// untyped: NULL !
		return nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			// untyped: NULL !
			return nil
		}
		// case *custompb.Lookup !!! above !
		rv = reflect.Indirect(rv)
	}
	switch rv.Kind() {
	// custom.LIST
	case reflect.Slice:
		{
			size := rv.Len()
			list := make([]any, size)
			for i := range size {
				list[i] = mapValue(
					rv.Index(i).Interface(),
				)
			}
			return list
		}
	// custom.[U]INT | FLOAT | ..
	default:
		{
			return rv.Interface()
		}
	}
	// // origin
	// return v
}

func (e *Record) AsMap() map[string]any {
	if e == nil {
		return nil
	}
	n := len(e.fields)
	m := make(map[string]any, n)
	_, cx := e.typeof.(customrel.ExtensionDescriptor)
	pk := e.typeof.Primary().Name()
	// [NOTE]: Populated ONLY !
	e.Range(func(fd customrel.FieldDescriptor, v any) bool {
		if v == nil {
			// No value = no output !
			return true
		}
		// Hide PK field for extension(s)..
		if cx && fd.Name() == pk {
			cx = false  // once
			return true // skip
		}
		v = mapValue(v)
		m[fd.Name()] = v
		return true
	})
	// if len(m) == 0 {
	// 	// No data !
	// 	return nil
	// }
	return m
}

func (e *Record) Proto() *structpb.Struct {
	src := e.AsMap()
	// if len(src) == 0 {
	if src == nil {
		// No data !
		return nil
	}
	obj, err := structpb.NewStruct(src)
	if err != nil {
		panic(fmt.Errorf("structpb.NewStruct(%#v); error: %v", src, err))
	}
	return obj
}

// FromProto decodes given [set] *Struct object into *Record fields
func (e *Record) FromProto(data *structpb.Struct) error {
	n := len(data.GetFields())
	if n == 0 {
		// empty !
		return nil
	}
	var (
		err    error
		typeof = e.typeof
		fields = typeof.Fields()
		field  customrel.FieldDescriptor
	)
	for name, value := range data.AsMap() {
		if field = fields.ByName(name); field == nil {
			return fmt.Errorf("record(%s).set(%s); no such field", typeof.Name(), name)
		}
		if err = e.Set(field, value); err != nil {
			return err
		}
	}
	return nil
}
