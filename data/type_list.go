package data

import (
	"fmt"
	"reflect"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// List of primitives type.
type List struct {
	elem customrel.Type
}

// ListAs primitive Type
func ListAs(elem Type) Type {
	if this, is := elem.(*List); is {
		return this
	}
	return &List{elem}
}

// Elem returns Type of element(s) in the List.
func (dt *List) Elem() customrel.Type {
	if dt != nil {
		return dt.elem
	}
	// UNDEFINED
	return nil
}

var _ customrel.Type = (*List)(nil)

// Kind of the data type.
func (*List) Kind() customrel.Kind {
	return customrel.LIST
}

// New data value codec.
func (dt *List) New() customrel.Codec {
	return &ListValue{
		typof: dt,
	}
}

var ErrListNoType = fmt.Errorf("custom: list[type] undefined")

// Err to check data type descriptor integrity.
func (dt *List) Err() error {
	// if dt == nil || dt.elem == nil {
	if dt == nil || dt.elem == nil {
		return ErrListNoType
	}
	return dt.elem.Err()
}

func (*List) Custom(pragma.DoNotImplement) {}

// ---------------------------------------- //
//             List of Value(s)             //
// ---------------------------------------- //

type ListValue struct {
	typof *List
	slice reflect.Value // any // []any
}

var _ customrel.Codec = (*ListValue)(nil)

// // Interface of the underlying value.
// // *int[32|64]
// // *uint[32|64]
// // *float[32|64]
// // *time.Time
// // *time.Duration
// // *Lookup
// // []any | LIST
// // map[string]any | RECORD
// func (dv *ListValue) Interface() any {
// 	panic("not implemented") // TODO: Implement
// }

// // Type of the data value
// func (dv *ListValue) Type() customrel.Type {
// 	panic("not implemented") // TODO: Implement
// }

// // Err to check underlying value
// //
// //	according to the data type constraints
// func (dv *ListValue) Err() error {
// 	panic("not implemented") // TODO: Implement
// }

// func newListValue(listOf *List) *ListValue {
// 	etyp := listOf.of
// 	eval := etyp.New()
// 	rtyp := reflect.TypeOf(eval)
// 	rval := reflect.New(reflect.SliceOf(rtyp))
// 	return &ListValue{
// 		typo: listOf,
// 		data: rval, // IsNil(!)
// 	}
// }

// Data type validation error
func (dv *ListValue) Err() error {
	// vs := v.data.Interface()
	// _ = vs
	if dv.IsNull() {
		return nil // NULL
	}
	n := dv.slice.Len()
	for i := 0; i < n; i++ {
		item := dv.slice.Index(i)
		want := dv.typof.elem.New()
		err := want.Decode(item.Interface())
		if err != nil {
			return err
		}
		err = want.Err()
		if err != nil {
			return err
		}
		// OK
	}
	return nil
}

// Type of the data value
func (dv *ListValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	return (*List)(nil) // err
}

// Interface Go value.
func (dv *ListValue) Interface() any {
	if !dv.IsNull() {
		return dv.slice.Interface()
	}
	return ([]any)(nil)
}

func (dv *ListValue) Decode(src any) error {

	if src == nil {
		dv.slice = reflect.Value{} // NULL
		return nil
	}

	// panic("not implemented") // TODO: Implement
	dtyp := dv.typof.elem
	etyp := dtyp.New().Interface()
	rtyp := reflect.TypeOf(etyp)
	// rtyp = reflect.SliceOf(rtyp)

	indirect := func(v any) (any, error) {
		switch v := v.(type) {
		case *structpb.Value:
			{
				switch v := v.GetKind().(type) {
				case *structpb.Value_NullValue:
					return nil, nil // NULL, OK
				case *structpb.Value_ListValue:
					return v.ListValue.GetValues(), nil
					// case *structpb.Value_BoolValue:
					// case *structpb.Value_NumberValue:
					// case *structpb.Value_StringValue:
					// case *structpb.Value_StructValue:
				}
				return nil, fmt.Errorf(
					"convert %T into *List[%s]",
					src, dtyp.Kind(),
				)
			}
		}
		return v, nil
	}
	src, err := indirect(src)
	if err != nil {
		return err
	}
	if src == nil {
		dv.slice = reflect.Value{} // NULL
		return nil
	}

	var n int
	srcVal := reflect.ValueOf(src)
	switch srcVal.Kind() {
	// *structpb.Value e.g.: structpb.Value{List:[]}
	// case reflect.Pointer:
	case reflect.Slice, reflect.Array:
		{
			n = srcVal.Len()
		}
	default:
		return fmt.Errorf(
			"convert %T into *List[%s]",
			src, dv.typof.elem.Kind(),
		)
	}

	// Iterate thru input(src) slice items
	rval := reflect.MakeSlice(
		reflect.SliceOf(rtyp),
		0, n,
	)
	for i := 0; i < n; i++ {
		srcE := srcVal.Index(i)
		dstE := dv.typof.elem.New()
		err := dstE.Decode(srcE.Interface())
		if err != nil {
			// Invalid input item.(value) spec.
			// according to the List[Type] descriptor
			return err
		}
		rval = reflect.Append(
			rval, reflect.ValueOf(
				dstE.Interface(),
			),
		)
	}

	dv.slice = rval
	return nil
}

func (dv *ListValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

// func (dv *ListValue) Compare(v2 Value) Match {
// 	panic("not implemented") // TODO: Implement
// }

func (dv *ListValue) IsNull() bool {
	return dv == nil || !dv.slice.IsValid() || dv.slice.IsNil()
}

func (dv *ListValue) IsZero() bool {
	return !dv.IsNull() && dv.slice.IsZero() // dv.slice.Len() == 0
}

func (*ListValue) Custom(pragma.DoNotImplement) {}
