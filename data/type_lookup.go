package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	customreg "github.com/webitel/custom/registry"
	custompb "github.com/webitel/proto/gen/custom"
	datapb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	// customregistry "webitel.go/service/custom/registry"
)

// Lookup reference type descriptor
type Lookup struct {
	spec *datapb.Lookup
	rel  customrel.DictionaryDescriptor
	err  error
}

var errLookupNoType = fmt.Errorf("lookup( type: ! ) required")

func ErrLookupPath(typeOf string) error {
	if typeOf == "" {
		return fmt.Errorf("lookup( path: ! ) required")
	}
	return fmt.Errorf("lookup( path: %q ) not found", typeOf)
}

type DictionaryTypeResolveFunc func(ctx context.Context, dc int64, pkg string) (customrel.DictionaryDescriptor, error)

// LookupAs primitive reference type
func LookupAs(
	ctx context.Context, dc int64,
	spec *datapb.Lookup, find ...DictionaryTypeResolveFunc,
) Type {

	ref := &Lookup{
		spec: spec,
	}
	if spec.GetPath() == "" {
		ref.err = ErrLookupPath("")
		return ref // .Err(!)
	}
	// [FIXME] ..
	if ctx == nil {
		ctx = context.TODO()
	}
	// Resolve type of relation ...
	for _, find := range find {
		ref.rel, ref.err = find(
			ctx, dc, spec.Path,
		)
		if ref.rel != nil {
			// ref.err = nil
			break // Found !
		}
	}
	// Finally ...
	if ref.rel == nil {
		ref.rel, ref.err = customreg.GetDictionary(
			ctx, dc, spec.Path,
		)
		if ref.rel == nil {
			// invalid lookup.type spec
			ref.err = ErrLookupPath(spec.Path)
			return ref
		}
	}
	return ref
}

// Dictionary as a related data type
func (dt *Lookup) Dictionary() customrel.DictionaryDescriptor {
	if dt != nil {
		return dt.rel
	}
	return nil
}

var _ customrel.Type = (*Lookup)(nil)

// Kind of basic type
// Type() Type // indirect type for: list, lookup
func (*Lookup) Kind() customrel.Kind {
	return customrel.LOOKUP
}

// With type constraints
// Sub(typ any) Type
// New value of this type
func (dt *Lookup) New() customrel.Codec {
	return &LookupValue{
		typof: dt,
	}
}

// Err of the related Dictionary type resolution.
func (dt *Lookup) Err() error {
	if dt.err != nil {
		return dt.err
	}
	return nil
}

// pragma.DoNotImplement
func (*Lookup) Custom(pragma.DoNotImplement) {}

// reference as a Lookup compatible Value.
type reference interface {
	// [Required] GetId value of the field, marked it's type structure as `primary`.
	GetId() string
	// [Readonly] GetName returns display, user-friendly name of the data record.
	GetName() string
	// [Optional] GetType SHOULD return the value of its record.[type].path
	// to help uniquely identify the type of the record's data structure
	// if it is not known ahead of time.
	// Optional: depends on context of operation.
	GetType() string
}

// Lookup type value codec
type LookupValue struct {
	typof *Lookup
	value *custompb.Lookup
	// id     any
}

var _ customrel.Codec = (*LookupValue)(nil)

// Data type validation error
func (dv *LookupValue) Err() error {
	if dv.IsNull() {
		// No value to check !
		return nil
	}
	// [CHECK]: v.typeOf.desc.(*pbtype.Lookup).Type().Primary()
	ref := dv.typof
	rel := ref.Dictionary()
	att := rel.Primary()
	val := att.Type().New()
	err := val.Decode(dv.value.GetId())
	if err != nil {
		return err
	}
	// REQUIRED
	// if val.IsNull() {
	if customrel.IsNull(val) {
		return fmt.Errorf("lookup( id: ! ); required")
	}
	// _ = val.Interface()
	return nil
}

// Type of the Value
func (dv *LookupValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	return (*Lookup)(nil) // err
}

// Cast(Type) Value
// Interface Go value.
func (dv *LookupValue) Interface() any {
	if !dv.IsNull() {
		return dv.value
	}
	return (*custompb.Lookup)(nil)
}

// func (v *LookupValue) Interface() any {
// 	if v != nil {
// 		return v.value
// 	}
// 	return (*pbdata.LookupValue)(nil)
// }

func (dv *LookupValue) Decode(src any) error {
	// panic("not implemented") // TODO: Implement
	// accept: src.(type)
	if src == nil {
		dv.value = nil // NULL
		return nil
	}
	// with .typeOf constraints
	typeOf := dv.Type().(*Lookup)
	// type has valid descriptor ?
	if err := typeOf.Err(); err != nil {
		return err
	}
	setValue := func(val *custompb.Lookup) error {
		if val == nil {
			// NULL
			dv.value = nil
			return nil
		}
		if val.Id == "" &&
			val.Name == "" &&
			val.Type == "" {
			// NULL
			dv.value = nil
			return nil
		}
		// Require: [ref.id]
		if val.Id == "" {
			// Reference object primary key required !
			return fmt.Errorf("lookup.id required")
		}
		// Check PK typ value
		pk := dv.typof.Dictionary().Primary()
		id := pk.Type().New()
		err := id.Decode(val.Id)
		if err != nil {
			return fmt.Errorf("lookup( id: %s ); %v", val.Id, err)
		}
		dv.value = val
		return nil // OK
	}
	stringValue := func(src *string) error {
		if src == nil {
			return setValue(nil)
		}
		input := strings.TrimSpace(*src)
		if input == "" {
			return setValue(nil)
		}
		// value, err := strconv.ParseInt(input, 10, int(typeOf.Bits()))
		// if err != nil {
		// 	return fmt.Errorf("convert: string %q value into lookup reference")
		// }
		value := &custompb.Lookup{
			Id: input,
		}
		return setValue(value)
	}
	protobufValue := func(src *structpb.Value) error {
		if src == nil {
			return setValue(nil)
		}
		switch input := src.GetKind().(type) {
		case *structpb.Value_NullValue:
			{
				return setValue(nil)
			}
		case *structpb.Value_NumberValue:
			{
				// if input == nil {
				// 	v.value = nil // NULL
				// 	break
				// }
				dv.value = &custompb.Lookup{
					Id: strconv.FormatFloat(input.NumberValue, 'f', -1, 64),
				}
			}
		case *structpb.Value_StringValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				return stringValue(&input.StringValue)
			}
		// case *structpb.Value_BoolValue:
		case *structpb.Value_StructValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				obj := input.StructValue
				if len(obj.GetFields()) == 0 {
					dv.value = nil // NULL
					break
				}
				value := &custompb.Lookup{}
				for h, v := range obj.Fields {
					switch strings.ToLower(h) {
					case "id":
						{
							if v := v.AsInterface(); v != nil {
								value.Id = fmt.Sprintf("%v", v)
							}
						}
					case "name":
						{
							if v := v.AsInterface(); v != nil {
								value.Name = fmt.Sprintf("%v", v)
							}
						}
					case "type":
						{
							if v := v.AsInterface(); v != nil {
								value.Type = fmt.Sprintf("%v", v)
							}
						}
					default:
						{
							return fmt.Errorf(
								"convert: %v value into Lookup reference",
								obj.AsMap(),
							)
						}
					}
				}
				setValue(value)
			}
		// case *structpb.Value_ListValue:
		default:
			{
				ref := src.ProtoReflect()
				def := ref.Descriptor()
				fd := def.Fields().ByName("kind")
				return fmt.Errorf(
					"convert: %s %v value into *Lookup reference", strings.TrimSuffix(string(
						// ref.WhichOneof(def.Oneofs().ByName("kind")).Name()),
						ref.WhichOneof(fd.ContainingOneof()).Name()),
						"_value",
					), ref.Get(fd).String(),
				)
			}
		}
		return nil
	}
	switch input := src.(type) {
	case *custompb.Lookup:
		{
			return setValue(input)
		}
	case *LookupValue:
		{
			return setValue(input.value)
		}
	// AS ${.id}
	case int32:
	case *int32:
	case int64:
		{
			return setValue(&custompb.Lookup{
				Id: strconv.FormatInt(input, 10),
			})
		}
	case *int64:
	case int, *int:
	case uint, *uint:
		//
	case int8, *int8:
	case int16, *int16:
		//
	case uint8, *uint8:
	case uint16, *uint16:
	case uint32, *uint32:
	case uint64, *uint64:

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
	case *wrapperspb.Int64Value:
	case *wrapperspb.UInt64Value:
	case *wrapperspb.Int32Value:
	case *wrapperspb.UInt32Value:
	case *wrapperspb.FloatValue:
	case *wrapperspb.DoubleValue:
	case *wrapperspb.StringValue:
	default:
		{
			return fmt.Errorf(
				"convert: %[1]T %[1]v value into *Lookup reference",
				src,
			)
		}
	}
	return nil // OK
}

func (dv *LookupValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

// func (dv *LookupValue) Compare(v2 Value) Match {
// 	panic("not implemented") // TODO: Implement
// }

func (dv *LookupValue) IsNull() bool {
	return dv == nil || dv.value.GetId() == "" // v.value.GetId() == ""
}

func (dv *LookupValue) IsZero() bool {
	return false
}

// pragma.DoNotImplement
func (*LookupValue) Custom(pragma.DoNotImplement) {}
