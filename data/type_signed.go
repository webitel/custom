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
type Signed struct {
	bits uint8       // [ 8, 16, 32, 64 ]
	spec *datapb.Int // descriptor: constraints
	min  int64
	max  int64
	err  error // .spec.(constraints) failed
}

// Integer based types.
var (
	Int8 = Signed{
		bits: 8,
		min:  -1 << 7,  // math.MinInt8,
		max:  1<<7 - 1, // math.MaxInt8,
		// violation: struct{min string; max string}{},
	}

	Int16 = Signed{
		bits: 16,
		min:  -1 << 15,  // math.MinInt16,
		max:  1<<15 - 1, // math.MaxInt16,
		// violation: struct{min string; max string}{},
	}

	Int32 = Signed{
		bits: 32,
		min:  -1 << 31,  // math.MinInt32,
		max:  1<<31 - 1, // math.MaxInt32,
		// violation: struct{min string; max string}{},
	}

	Int64 = Signed{
		bits: 64,
		min:  -1 << 63,  // math.MinInt64,
		max:  1<<63 - 1, // math.MaxInt64,
		// violation: struct{min string; max string}{},
	}

	Int = Int32
)

// SignedOf integer type.
func signedAs(bitsize uint8, spec *datapb.Int) *Signed {
	integer := &Signed{
		bits: bitsize,
		spec: spec,
	}
	integer.setup()
	return integer
}

// SignedAs integer type.
func SignedAs(bitsize uint8, spec *datapb.Int) Type {
	return signedAs(bitsize, spec)
}

var (
	// signed base types
	signedBits = map[uint8]*Signed{
		8:  &Int8,
		16: &Int16,
		32: &Int32,
		64: &Int64,
	}
)

// setup integer type value boundaries
func (dt *Signed) setup() (err error) {
	defer func() {
		dt.err = err
	}()
	base := signedBits[dt.bits]
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
func (dt *Signed) accept(arg string, v int64) (err error) {
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

// As [dt] signed integer base type with custom constraints.
func (dt *Signed) As(spec *datapb.Int) *Signed {
	if dt.spec == spec {
		return dt // this
	}
	return signedAs(dt.bits, spec)
}

// Bits size of signed integer type value.
func (dt *Signed) Bits() uint8 {
	if dt != nil {
		return dt.bits
	}
	return 0 // as 32
}

// MinValue constraint of this signed integer type.
func (dt *Signed) MinValue() int64 {
	return dt.min
}

// MinValue constraint of this signed integer type.
func (dt *Signed) MaxValue() int64 {
	return dt.max
}

var _ customrel.Type = (*Signed)(nil)

func (dt *Signed) Kind() customrel.Kind {
	if dt != nil {
		switch dt.bits {
		case 32:
			return customrel.INT32
		case 64:
			return customrel.INT64
		}
	}
	// default
	return customrel.INT
}

func (dt *Signed) New() customrel.Codec {
	return &SignedValue{
		typof: dt,
	}
}

func (dt *Signed) Err() error {
	if dt != nil {
		return dt.err
	}
	return nil
}

func (*Signed) Custom(pragma.DoNotImplement) {}

// SignedValue represents an integer value
type SignedValue struct {
	typof *Signed
	value *int64 // NULL(-able)
}

var _ customrel.Codec = (*SignedValue)(nil)

// Interface of the GoValue
func (dv *SignedValue) Interface() any {
	if dv != nil {
		return dv.value
	}
	// typical: NULL
	return (*int64)(nil)
}

func (dv *SignedValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	// default: INT
	return (*Signed)(nil)
}

func (dv *SignedValue) Err() error {
	if dv.IsNull() {
		return nil
	}
	return dv.typof.accept(" ", *dv.value)
}

// implements [Nullable] interface
func (dv *SignedValue) IsNull() bool {
	return dv == nil || dv.value == nil
}

func (dv *SignedValue) IsZero() bool {
	return !dv.IsNull() && (*dv.value) == 0
}

func (dv *SignedValue) Decode(src any) error {
	// accept: src.(type)
	if src == nil {
		dv.value = nil // NULL
		return nil
	}
	// with .typeOf constraints
	typeOf := dv.Type().(*Signed)
	setValue := func(val *int64) error {
		if val == nil {
			dv.value = nil
			return nil // NULL
		}
		err := typeOf.accept(" ", int64(*val))
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
		value, err := strconv.ParseInt(input, 10, int(typeOf.Bits()))
		if err != nil {
			return fmt.Errorf("convert: string %q value into signed integer", input)
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
					return v == float64(int64(v))
				}
				if !integral(input.NumberValue) {
					return fmt.Errorf(
						"convert: float number %v into int32",
						input.NumberValue,
					)
				}
				value := int64(input.NumberValue)
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
	switch input := src.(type) {
	case SignedValue:
		{
			return setValue(input.value)
		}
	case int32:
		{
			value := int64(input)
			return setValue(&value)
		}
	case *int32:
		{
			if input == nil {
				return setValue(nil)
			}
			value := int64(*input)
			return setValue(&value)
		}
	case int64:
		{
			err := typeOf.accept(" ", input)
			if err != nil {
				return err
			}
			value := input
			dv.value = &value
		}
	case *int64:
		{
			if input == nil {
				return setValue(nil)
			}
			err := typeOf.accept(" ", *input)
			if err != nil {
				return err
			}
			value := int64(*input)
			dv.value = &value
		}
		//
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
				"convert: %[1]T %[1]v value into int32",
				src,
			)
		}
	}
	return nil // OK
}

func (dv *SignedValue) Encode(dst any) error {
	panic("not implemented") // TODO: Implement
}

func (*SignedValue) Custom(pragma.DoNotImplement) {}
