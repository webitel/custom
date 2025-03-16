package customrel

import (
	"reflect"

	"github.com/webitel/custom/internal/pragma"
)

// Codec for the data type value convertions
type Codec interface {
	// Interface of the underlying value.
	// *int[32|64]
	// *uint[32|64]
	// *float[32|64]
	// *time.Time
	// *time.Duration
	// *Lookup
	// []any | LIST
	// map[string]any | RECORD
	Interface() any

	Decode(src any) error
	Encode(dst any) error

	// Type of the data value
	Type() Type

	// Err to check underlying value
	//  according to the data type constraints
	Err() error

	pragma.DoNotImplement
}

type Nullable interface {
	IsNull() bool
}

func IsNull(v any) bool {
	if v == nil {
		return true
	}
	if self, is := v.(Nullable); is {
		return self.IsNull()
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.UnsafePointer,
		reflect.Interface,
		reflect.Pointer,
		reflect.Slice,
		reflect.Chan,
		reflect.Func,
		reflect.Map:
		{
			return rv.IsNil()
		}
	}
	return false
}

// Indirect converts custom NULL(-able) value to Go(-style).
// MAY return:
// nil
// int[32|64]
// uint[32|64]
// float[32|64]
// time.Time
// time.Duration
// *Lookup
// []any | LIST
// map[string]any | RECORD
func Indirect(v any) any {
	panic("not implemented")
}
