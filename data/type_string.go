package data

import (
	"fmt"
	"strings"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	datapb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type String struct {
	spec *datapb.Text
	err  error
}

// StringAs primtive type
func StringAs(spec *datapb.Text) Type {
	return &String{spec: spec}
}

var _ Type = (*String)(nil)

// Kind of the data type.
func (dt *String) Kind() customrel.Kind {
	return customrel.STRING
}

// New data value codec.
func (dt *String) New() customrel.Codec {
	return &StringValue{typof: dt}
}

// Err to check data type descriptor integrity.
func (dt *String) Err() error {
	if dt != nil {
		return dt.err
	}
	return nil
}

func (dt *String) Custom(pragma.DoNotImplement) {}

type StringValue struct {
	typof *String
	value *string
}

var _ customrel.Codec = (*StringValue)(nil)

// Interface of the underlying value.
// *int[32|64]
// *uint[32|64]
// *float[32|64]
// *time.Time
// *time.Duration
// *Lookup
// []any | LIST
// map[string]any | RECORD
func (dv *StringValue) Interface() any {
	return dv.value // (*string)(nil)
}

func (dv *StringValue) Decode(src any) error {
	// accept: src.(type)
	if src == nil {
		dv.value = nil // NULL
		return nil
	}
	// with .typeOf constraints
	// typeOf := v.Type().(*String)
	setValue := func(val *string) error {
		if val == nil {
			dv.value = nil
			return nil // NULL
		}
		dv.value = val
		return nil // OK
	}
	protobufValue := func(src *structpb.Value) error {
		if src == nil {
			// NULL
			dv.value = nil
			return nil
		}
		switch input := src.GetKind().(type) {
		case *structpb.Value_NullValue:
			{
				dv.value = nil // NULL
			}
		case *structpb.Value_StringValue:
			// case *structpb.Value_NumberValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				return setValue(&input.StringValue)
			}
		// case *structpb.Value_BoolValue:
		// case *structpb.Value_StructValue:
		// case *structpb.Value_ListValue:
		default:
			{
				ref := src.ProtoReflect()
				def := ref.Descriptor()
				// fd := def.Fields().ByName("kind")
				kind := def.Oneofs().ByName("kind")
				value := ref.WhichOneof(kind)
				return fmt.Errorf(
					"convert: %s %v value into String", strings.TrimSuffix(string(
						// ref.WhichOneof(def.Oneofs().ByName("kind")).Name()),
						// ref.WhichOneof(fd.ContainingOneof()).Name()),
						value.Name()),
						"_value",
					), ref.Get(value).String(),
				)
			}
		}
		return nil
	}
	switch input := src.(type) {
	case StringValue:
		{
			return setValue(input.value)
		}
	// case int32:
	// case *int32:
	// case int64:
	// case *int64:
	// 	//
	// case int, *int:
	// case uint, *uint:
	// 	//
	// case int8, *int8:
	// case int16, *int16:
	// 	//
	// case uint8, *uint8:
	// case uint16, *uint16:
	// case uint32, *uint32:
	// case uint64, *uint64:

	case string:
		{
			return setValue(&input)
		}
	case *string:
		{
			return setValue(input)
		}
	case *structpb.Value:
		{
			return protobufValue(input)
		}
	// case *wrapperspb.Int64Value:
	// case *wrapperspb.UInt64Value:
	// case *wrapperspb.Int32Value:
	// case *wrapperspb.UInt32Value:
	// case *wrapperspb.FloatValue:
	// case *wrapperspb.DoubleValue:
	case *wrapperspb.StringValue:
		{
			if input == nil {
				dv.value = nil
				break // NULL
			}
			return setValue(&input.Value)
		}
	default:
		{
			return fmt.Errorf(
				"convert: %[1]T %[1]v value into String",
				src,
			)
		}
	}
	return nil // OK
}

func (dv *StringValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

// Type of the data value
func (dv *StringValue) Type() customrel.Type {
	panic("not implemented") // TODO: Implement
}

// Err to check underlying value
//
//	according to the data type constraints
func (dv *StringValue) Err() error {
	panic("not implemented") // TODO: Implement
}

func (*StringValue) Custom(pragma.DoNotImplement) {}
