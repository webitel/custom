package data

import (
	"reflect"

	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
)

type (
	// Primitive data type
	// github.com/webitel/custom/reflect.Type
	Type = customrel.Type
)

// Elem returns the undelying element type.
// For given (*List) returns it's element type.
// Otherwise returns given type as a result.
func Elem(of Type) Type {
	if list, is := of.(*List); is {
		return list.Elem()
	}
	return of
}

// Indirect casts given [t] type of [v] value
// to it's Go native value equivalent.
//
// <nil>
// int64
// float64
// string
// time.Time
// Duration
// []any
// map[string]any
func Indirect(t Type, v any) (any, error) {
	if v == nil {
		// NULL
		return nil, nil
	}
	var err error
	switch e := v.(type) {
	case *custompb.Lookup:
		{
			if e.GetId() == "" {
				// NULL
				return nil, nil
			}
			// switch fd.Kind() {}
			rt := t.(*Lookup)               // reference.(lookup) type
			pk := rt.Dictionary().Primary() // reference.[primary] field
			rv := pk.Type().New()           // value type codec
			err = rv.Decode(e.GetId())
			if err != nil {
				// failed to cast string value to it's data type !
				return e.GetId(), err // string value !
			}
			// return rv.Interface(), nil // [primary] key type Go value !
			v = rv.Interface() // NULL(-able) !
			// process as indirect below !
		}
	}
	// reflect
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		// NULL !
		return nil, nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			// NULL !
			return nil, nil
		}
		// *int64
		// *uint64
		// *tim.Time
		// ...
		rv = reflect.Indirect(rv)
	}
	switch rv.Kind() {
	case reflect.Slice:
		{
			listT := t.(*List)
			itemT := listT.Elem()

			sizeV := rv.Len()
			listV := make([]any, sizeV)

			for i := 0; i < sizeV; i++ {
				listV[i], err = Indirect(
					itemT, rv.Index(i).Interface(),
				)
				if err != nil {
					// failed to cast list[item] value !
					return nil, err
				}
			}
			return listV, nil
		}
	}
	// [primitive] !
	return rv.Interface(), nil
}
