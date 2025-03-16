package data

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	datapb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Boolean struct {
	spec *datapb.Bool
}

func BoolAs(spec *datapb.Bool) Type {
	return &Boolean{spec: spec}
}

var _ customrel.Type = (*Boolean)(nil)

func (*Boolean) Custom(pragma.DoNotImplement) {}

// Kind of the data type.
func (*Boolean) Kind() customrel.Kind {
	return customrel.BOOL
}

// New data value codec.
func (dt *Boolean) New() customrel.Codec {
	return &BoolValue{typof: dt}
}

// Err to check data type descriptor integrity.
func (dt Boolean) Err() error {
	return nil
}

type BoolValue struct {
	typof *Boolean
	value *bool
}

var _ customrel.Codec = (*BoolValue)(nil)

// Interface of the *bool value.
func (dv *BoolValue) Interface() any {
	if dv.value != nil {
		return dv.value
	}
	return (*bool)(nil)
}

func (dv *BoolValue) Decode(src any) error {
	// accept: src.(type)
	if src == nil {
		dv.value = nil // NULL
		return nil
	}
	stringValue := func(src *string) error {
		if src == nil {
			dv.value = nil // NULL
			return nil
		}
		input := strings.TrimSpace(*src)
		if input == "" {
			dv.value = nil // NULL
			return nil
		}
		value, err := strconv.ParseBool(input)
		if err != nil {
			return fmt.Errorf("convert: string %q value into bool", input)
		}
		dv.value = &value
		return nil
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
		case *structpb.Value_BoolValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				dv.value = &input.BoolValue
			}
		case *structpb.Value_StringValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				return stringValue(&input.StringValue)
			}
		// case *structpb.Value_NumberValue:
		// case *structpb.Value_StructValue:
		// case *structpb.Value_ListValue:
		default:
			{
				ref := src.ProtoReflect()
				def := ref.Descriptor()
				fd := def.Fields().ByName("kind")
				return fmt.Errorf(
					"convert: %s %v value into bool", strings.TrimSuffix(string(
						// ref.WhichOneof(def.Oneofs().ByName("kind")).Name()),
						ref.WhichOneof(fd.ContainingOneof()).Name()),
						"_value",
					), ref.Get(fd).String(),
				)
			}
		}
		return nil
	}
	protobufBoolValue := func(src *wrapperspb.BoolValue) error {
		if src == nil {
			// NULL
			dv.value = nil
			return nil
		}
		dv.value = &src.Value
		return nil
	}
	switch input := src.(type) {
	case BoolValue:
		{
			// internal
			dv.value = input.value
		}
	case bool:
		{
			dv.value = &input
		}
	case *bool:
		{
			dv.value = input
		}
	case string:
		{
			return stringValue(&input)
		}
	case *string:
		{
			return stringValue(input)
		}
	case *structpb.Value:
		{
			return protobufValue(input)
		}
	case **structpb.Value:
		{
			if input == nil {
				dv.value = nil
				break // NULL
			}
			return protobufValue(*input)
		}
	case *wrapperspb.BoolValue:
		{
			return protobufBoolValue(input)
		}
	case **wrapperspb.BoolValue:
		{
			if input == nil {
				dv.value = nil
				break // NULL
			}
			return protobufBoolValue(*input)
		}
	default:
		{
			return fmt.Errorf(
				"convert: %[1]T %[1]v value into bool",
				src,
			)
		}
	}
	return nil // OK
}

func (dv *BoolValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

// Type of the data value
func (dv *BoolValue) Type() customrel.Type {
	return dv.typof
}

// Err to check underlying value
//
//	according to the data type constraints
func (dv *BoolValue) Err() error {
	// nil | false | true
	return nil
}

func (*BoolValue) Custom(pragma.DoNotImplement) {}
