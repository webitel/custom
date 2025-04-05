package data

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
func (*String) Kind() customrel.Kind {
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

func (*String) Custom(pragma.DoNotImplement) {}

var stringViolations = map[string]string{
	"multiline": "string {{.value}} violates multiline constraint",
	"max_bytes": "string {{.value}} violates max boundary of: {{.max_bytes}} bytes",
	"max_chars": "string {{.value}} violates max boundary of: {{.max_chars}} characters",
}

func (dt *String) violationError(kind string, val *string) error {
	text := (*val)
	if num := utf8.RuneCountInString(text); num > 16 {
		text = fmt.Sprintf("%s..(+%d)", []byte(text)[0:14], (num - 16))
	}
	tmpl := stringViolations[kind]
	tmpl = strings.ReplaceAll(tmpl, "{{.value}}", fmt.Sprintf("%q", text))
	tmpl = strings.ReplaceAll(tmpl, "{{.max_bytes}}", fmt.Sprintf("%d", dt.spec.GetMaxBytes()))
	tmpl = strings.ReplaceAll(tmpl, "{{.max_chars}}", fmt.Sprintf("%d", dt.spec.GetMaxChars()))
	return RequestError(
		fmt.Sprintf("custom.type.string.%s.violation", kind), ("custom: " + tmpl),
	)
}

// Accept typical value type constraints
func (dt *String) Accept(val *string) error {
	if dt == nil {
		// no constraints
		return nil // [OK] ; whatever ..
	}
	if dt.err != nil {
		// invalid type descriptor
		return dt.err
	}
	if val == nil || *val == "" {
		return nil // [OK] ; NULL or empty
	}
	if !utf8.ValidString(*val) {
		return RequestError(
			"custom.type.string.encoding.error",
			"custom: string contains invalid UTF-8-encoded runes",
		)
	}
	if dt.spec == nil {
		// no constraints
		return nil // [OK] ; whatever ..
	}
	if !dt.spec.Multiline && strings.ContainsAny(*val, "\n\r") {
		return dt.violationError("multiline", val)
	}
	if max := dt.spec.MaxBytes; 0 < max && int(max) < len(*val) {
		return dt.violationError("max_bytes", val)
	}
	if max := dt.spec.MaxChars; 0 < max && int(max) < utf8.RuneCountInString(*val) {
		return dt.violationError("max_chars", val)
	}
	return nil // OK
}

type StringValue struct {
	typof *String
	value *string
}

var _ customrel.Codec = (*StringValue)(nil)

// Interface of the [*string] value.
func (dv *StringValue) Interface() any {
	return dv.value // (*string)(nil)
}

func (dv *StringValue) Decode(src any) error {
	setValue := func(set *string) (err error) {
		err = dv.typof.Accept(set)
		if err == nil {
			dv.value = set
		}
		return // err // ?
	}
	// accept: src.(type)
	if src == nil {
		return setValue(nil)
	}
	switch data := src.(type) {
	case StringValue:
		{
			return setValue(data.value)
		}
	case *StringValue:
		{
			if data == nil {
				return setValue(nil)
			}
			if data == dv {
				return nil // SELF
			}
			return setValue(data.value)
		}
	case string:
		{
			return setValue(&data)
		}
	case *string:
		{
			return setValue(data)
		}
	case *structpb.Value:
		{
			if data == nil {
				return setValue(nil)
			}
			switch kind := data.Kind.(type) {
			case nil:
				{
					return setValue(nil)
				}
			case *structpb.Value_NullValue:
				{
					return setValue(nil) // NULL
				}
			case *structpb.Value_NumberValue:
				{
					// if kind == nil {
					// 	return setValue(nil) // NULL
					// }
					value := fmt.Sprintf("%f", kind.NumberValue)
					return setValue(&value)
				}
			case *structpb.Value_StringValue:
				{
					// if kind == nil {
					// 	return setValue(nil) // NULL
					// }
					value := kind.StringValue
					return setValue(&value)
				}
			case *structpb.Value_BoolValue:
				{
					// if kind == nil {
					// 	return setValue(nil) // NULL
					// }
					value := fmt.Sprintf("%t", kind.BoolValue)
					return setValue(&value)
				}
			// case *structpb.Value_ListValue:
			// case *structpb.Value_StructValue:
			default:
				{
					ref := data.ProtoReflect()
					def := ref.Descriptor()
					fd := def.Fields().ByName("kind")
					return fmt.Errorf(
						"convert: %s value %v into String", strings.TrimSuffix(string(
							// ref.WhichOneof(def.Oneofs().ByName("kind")).Name()),
							ref.WhichOneof(fd.ContainingOneof()).Name()),
							"_value",
						), ref.Get(fd).String(),
					)
				}
			}
		}
	case *wrapperspb.Int64Value:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%d", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.UInt64Value:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%d", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.Int32Value:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%d", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.UInt32Value:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%d", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.FloatValue:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%f", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.DoubleValue:
		{
			if data == nil {
				return setValue(nil)
			}
			value := fmt.Sprintf("%f", data.Value)
			return setValue(&value)
		}
	case *wrapperspb.StringValue:
		{
			if data == nil {
				return setValue(nil)
			}
			value := data.Value
			return setValue(&value)
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

	default:
		{
			return fmt.Errorf(
				"convert: %[1]T value %[1]v into String",
				src,
			)
		}
	}
	panic("unreachable code")
	return nil // OK
}

func (dv *StringValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

// Type of the data value
func (dv *StringValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	return (*String)(nil)
}

// Err to check underlying value
//
//	according to the data type constraints
func (dv *StringValue) Err() error {
	// TODO: [re]design Codec interface method
	return nil
}

func (*StringValue) Custom(pragma.DoNotImplement) {}
