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

// Signed integer type
type Unsigned struct {
	bits uint8        // [ 8, 16, 32, 64 ]
	spec *datapb.Uint // descriptor: constraints
	min  uint64
	max  uint64
	err  error // .spec.(constraints) failed
}

// Integer based types.
var (
	Uint8 = Unsigned{
		bits: 8,
		min:  0,
		max:  1<<8 - 1, // math.MaxUint8,
		// violation: struct{min string; max string}{},
	}

	Uint16 = Unsigned{
		bits: 16,
		min:  0,
		max:  1<<16 - 1, // math.MaxUint16,
		// violation: struct{min string; max string}{},
	}

	Uint32 = Unsigned{
		bits: 32,
		min:  0,
		max:  1<<32 - 1, // math.MaxUint32,
		// violation: struct{min string; max string}{},
	}

	Uint64 = Unsigned{
		bits: 64,
		min:  0,
		max:  1<<64 - 1, // math.MaxUint64,
		// violation: struct{min string; max string}{},
	}

	Uint = Uint32
)

// UnsignedAs integer type.
func unsignedAs(bitsize uint8, spec *datapb.Uint) *Unsigned {
	integer := &Unsigned{
		bits: bitsize,
		spec: spec,
	}
	integer.setup()
	return integer
}

// UnsignedAs integer type.
func UnsignedAs(bitsize uint8, spec *datapb.Uint) Type {
	return unsignedAs(bitsize, spec)
}

var (
	// signed base types
	unsignedBits = map[uint8]*Unsigned{
		8:  &Uint8,
		16: &Uint16,
		32: &Uint32,
		64: &Uint64,
	}
)

// setup integer type value boundaries
func (dt *Unsigned) setup() (err error) {
	defer func() {
		dt.err = err
	}()
	base := unsignedBits[dt.bits]
	if base == nil {
		err = fmt.Errorf("int%d: invalid signed integer type", dt.bits)
		return
	}
	spec := dt.spec
	if spec == nil {
		dt.min = base.min
		dt.max = base.max
		return // nil
	}
	// // accept value `v` within base type boundaries
	// accept := func(op string, v int64) (ok bool) {
	// 	if v < base.min {
	// 		err = fmt.Errorf("int%d: invalid%svalue: %d; min: %d", base.bits, op, v, base.min)
	// 		return false // err
	// 	}
	// 	if base.max < v {
	// 		err = fmt.Errorf("int%d: invalid%svalue: %d; max: %d", base.bits, op, v, base.max)
	// 		return false // err
	// 	}
	// 	return true // ok
	// }
	min := base.min
	if set := spec.Min; set != nil {
		err = base.accept(" lower bound ", set.Value)
		if err != nil {
			return // err
		}
		min = set.Value
	}
	max := base.max
	if set := dt.spec.GetMax(); set != nil {
		err = base.accept(" upper bound ", set.Value)
		if err != nil {
			return // err
		}
		max = set.Value
	}
	if (max - min) < 1 {
		err = fmt.Errorf("int%d: invalid range of type values; min: %d; max: %d", base.bits, min, max)
		return // err
	}
	// apply
	dt.min = min
	dt.max = max
	return // nil // ok
}

// accept [arg]ument value [v] within type boundaries
func (dt *Unsigned) accept(arg string, v uint64) (err error) {
	n := len(arg)
	if n == 0 || arg[0] != ' ' {
		arg = " " + arg
		n++
	}
	if n > 1 && arg[n-1] != ' ' {
		arg += " "
	}
	if v < dt.min {
		err = fmt.Errorf(
			"int%d: invalid%svalue: %d; min: %d",
			dt.bits, arg, v, dt.min,
		)
		return // err
	}
	if dt.max < v {
		err = fmt.Errorf(
			"int%d: invalid%svalue: %d; max: %d",
			dt.bits, arg, v, dt.max,
		)
		return // err
	}
	return // nil
}

// As [dt] unsigned integer base type with custom constraints.
func (dt *Unsigned) As(spec *datapb.Uint) *Unsigned {
	if dt.spec == spec {
		return dt // this
	}
	return unsignedAs(dt.bits, spec)
}

// Bits size of signed integer type value.
func (dt *Unsigned) Bits() uint8 {
	if dt != nil {
		return dt.bits
	}
	return 0 // as 32
}

// MinValue constraint of this unsigned integer type.
func (dt *Unsigned) MinValue() uint64 {
	return dt.min
}

// MinValue constraint of this unsigned integer type.
func (dt *Unsigned) MaxValue() uint64 {
	return dt.max
}

var _ customrel.Type = (*Unsigned)(nil)

func (dt *Unsigned) Kind() customrel.Kind {
	if dt != nil {
		switch dt.bits {
		case 32:
			return customrel.UINT32
		case 64:
			return customrel.UINT64
		}
	}
	// default
	return customrel.UINT
}

func (dt *Unsigned) New() customrel.Codec {
	return &UnsignedValue{typof: dt}
}

func (dt *Unsigned) Err() error {
	if dt != nil {
		return dt.err
	}
	return nil
}

func (*Unsigned) Custom(pragma.DoNotImplement) {}

// SignedValue represents an integer value
type UnsignedValue struct {
	typof *Unsigned
	value *uint64 // NULL(-able)
}

var _ customrel.Codec = (*SignedValue)(nil)

// Interface of the GoValue
func (dv *UnsignedValue) Interface() any {
	if dv != nil {
		return dv.value
	}
	// typical: NULL
	return (*int64)(nil)
}

func (dv *UnsignedValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	// default: INT
	return (*Unsigned)(nil)
}

func (dv *UnsignedValue) Err() error {
	if dv.IsNull() {
		return nil
	}
	return dv.typof.accept(" ", *dv.value)
}

// implements [Nullable] interface
func (dv *UnsignedValue) IsNull() bool {
	return dv == nil || dv.value == nil
}

func (dv *UnsignedValue) IsZero() bool {
	return !dv.IsNull() && (*dv.value) == 0
}

func (dv *UnsignedValue) Decode(src any) error {
	// accept: src.(type)
	if src == nil {
		dv.value = nil // NULL
		return nil
	}
	// with .typeOf constraints
	typeOf := dv.Type().(*Unsigned)
	setValue := func(val *uint64) error {
		if val == nil {
			dv.value = nil
			return nil // NULL
		}
		err := typeOf.accept(" ", uint64(*val))
		if err != nil {
			return err
		}
		dv.value = val
		return nil // OK
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
		value, err := strconv.ParseUint(input, 10, int(typeOf.Bits()))
		if err != nil {
			return fmt.Errorf("convert: string %q value into unsigned integer", input)
		}
		return setValue(&value)
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
		case *structpb.Value_NumberValue:
			{
				if input == nil {
					dv.value = nil // NULL
					break
				}
				// https://stackoverflow.com/questions/43182427/how-to-check-if-float-value-is-actually-int
				integral := func(v float64) bool {
					return v == float64(uint64(v))
				}
				if !integral(input.NumberValue) {
					return fmt.Errorf(
						"convert: float number %v into uint32",
						input.NumberValue,
					)
				}
				value := uint64(input.NumberValue)
				dv.value = &value
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
		// case *structpb.Value_StructValue:
		// case *structpb.Value_ListValue:
		default:
			{
				ref := src.ProtoReflect()
				def := ref.Descriptor()
				fd := def.Fields().ByName("kind")
				return fmt.Errorf(
					"convert: %s %v value into unsigned integer", strings.TrimSuffix(string(
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
	case UnsignedValue:
		{
			return setValue(input.value)
		}
	// ------------------------- //
	case uint32:
		{
			value := uint64(input)
			return setValue(&value)
		}
	case *uint32:
		{
			if input == nil {
				return setValue(nil)
			}
			value := uint64(*input)
			return setValue(&value)
		}
	case uint64:
		{
			value := uint64(input)
			return setValue(&value)
		}
	case *uint64:
		{
			return setValue(input)
		}
	// ------------------------- //
	case int32:
		{
			if input < 0 {
				// negative integer !
				return fmt.Errorf("convert: int32(%d) value into uint", input)
			}
			value := uint64(input)
			return setValue(&value)
		}
	case *int32:
		{
			if input == nil {
				return setValue(nil)
			}
			if *input < 0 {
				// negative integer !
				return fmt.Errorf("convert: int32(%d) value into uint", input)
			}
			value := uint64(*input)
			return setValue(&value)
		}
	case int64:
		{
			if input < 0 {
				// negative integer !
				return fmt.Errorf("convert: int64(%d) value into uint", input)
			}
			value := uint64(input)
			return setValue(&value)
		}
	case *int64:
		{
			if input == nil {
				return setValue(nil)
			}
			if (*input) < 0 {
				// negative integer !
				return fmt.Errorf("convert: int64(%d) value into uint", input)
			}
			value := uint64(*input)
			return setValue(&value)
		}
	// ------------------------- //
	case int, *int:
	case uint, *uint:
		//
	case int8, *int8:
	case int16, *int16:
		//
	case uint8, *uint8:
	case uint16, *uint16:

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
				"convert: %[1]T %[1]v value into uint32",
				src,
			)
		}
	}
	return nil // OK
}

func (dv *UnsignedValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

func (*UnsignedValue) Custom(pragma.DoNotImplement) {}
