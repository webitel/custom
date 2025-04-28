package data

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	datapb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type DateTime struct {
	spec *datapb.Datetime
}

func DateTimeAs(spec *datapb.Datetime) Type {
	return &DateTime{spec: spec}
}

var _ Type = (*DateTime)(nil)

// Kind of the data type.
func (*DateTime) Kind() customrel.Kind {
	return customrel.DATETIME
}

// New data value codec.
func (dt *DateTime) New() customrel.Codec {
	return &DateTimeValue{typof: dt}
}

// Err to check data type descriptor integrity.
func (dt *DateTime) Err() error {
	// panic("not implemented")
	return nil
}

func (dt *DateTime) Custom(pragma.DoNotImplement) {}

// Format returns preconfigured input/output layout string
func (dt *DateTime) Format() string {
	if dt != nil {
		layout := dt.spec.GetFormat()
		if layout != "" {
			return layout
		}
	}
	// "2006-01-02 15:04:05"
	return time.DateTime
}

// Accept checks value type constraints
func (dt *DateTime) Accept(val *time.Time) error {
	if dt == nil || dt.spec == nil {
		// no constraints assigned !
		return nil // OK
	}
	// TODO:
	return nil // OK
}

type DateTimeValue struct {
	typof *DateTime
	value *time.Time
}

var _ customrel.Codec = (*DateTimeValue)(nil)

// Type of the data value
func (dv *DateTimeValue) Type() Type {
	if dv != nil {
		return dv.typof
	}
	return (*DateTime)(nil)
}

// Interface of the [*time.Time] value.
func (dv *DateTimeValue) Interface() any {
	if dv != nil {
		return dv.value
	}
	return (*time.Time)(nil)
}

// // DESIGN: Duration, Interval
// type Interval struct {
// 	Millisecond int64
// 	Days        int32
// 	Weeks       int32
// 	Months      int32
// }

func (dv *DateTimeValue) Decode(src any) error {
	setValue := func(set *time.Time) (err error) {
		err = dv.typof.Accept(set)
		if err != nil {
			return // [ERR]
		}
		dv.value = set
		return nil // [OK]
	}
	if src == nil {
		return setValue(nil)
	}
	data := reflect.ValueOf(src)
	if !data.IsValid() {
		return setValue(nil)
	}
	if data.Kind() == reflect.Pointer && data.IsNil() {
		return setValue(nil)
	}
	setInteger := func(set *int64) error {
		if set == nil {
			return setValue(nil)
		}
		// timestamp (seconds)
		date := time.Unix(*set, 0).UTC()
		return setValue(&date)
	}
	setDouble := func(set *float64) error {
		if set == nil {
			return setValue(nil)
		}
		date := CastNumberAsDateTime(*set, time.Millisecond)
		return setValue(&date)
	}
	setString := func(set *string) error {
		if set == nil {
			return setValue(nil)
		}
		var (
			err  error
			date time.Time
			text = strings.TrimSpace(*set)
		)
		if text == "" {
			return setValue(&date) // Zero(0)
		}
		for _, layout := range []string{
			// Default: Go time.String() format first !
			// https://cs.opensource.google/go/go/+/refs/tags/go1.23.3:src/time/format.go;l=545
			"2006-01-02 15:04:05.999999999 -0700 MST",
			dv.typof.Format(),
			time.RFC1123Z,
			time.RFC3339,
			time.RFC3339Nano,
			time.DateTime,
		} {
			date, err = time.Parse(layout, text)
			if err == nil {
				return setValue(&date)
			}
		}
		// finally: try to decode as timestamp[.ms] number
		if ts, err := strconv.ParseFloat(text, 64); err == nil {
			return setDouble(&ts)
		}
		return fmt.Errorf(
			"convert: string %q value into DateTime",
			*set,
		)
	}
	switch data := src.(type) {
	case time.Time:
		{
			return setValue(&data)
		}
	case *time.Time:
		{
			return setValue(data)
		}
	case int64:
		{
			return setInteger(&data)
		}
	case *int64:
		{
			return setInteger(data)
		}
	case float64:
		{
			return setDouble(&data)
		}
	case *float64:
		{
			return setDouble(data)
		}
	case string:
		{
			return setString(&data)
		}
	case *string:
		{
			return setString(data)
		}
	// protobuf .well-known
	case *structpb.Value:
		{
			// if input == nil {
			// 	return setValue(nil)
			// }
			switch kind := data.GetKind().(type) {
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
					// if input == nil {
					// 	return setValue(nil) // NULL
					// }
					value := kind.NumberValue
					return setDouble(&value)
				}
			case *structpb.Value_StringValue:
				{
					// if input == nil {
					// 	return setValue(nil) // NULL
					// }
					value := kind.StringValue
					return setString(&value)
				}
			// case *structpb.Value_BoolValue:
			// case *structpb.Value_ListValue:
			// case *structpb.Value_StructValue:
			default:
				{
					ref := data.ProtoReflect()
					def := ref.Descriptor()
					fd := def.Fields().ByName("kind")
					return fmt.Errorf(
						"convert: %s %v value into DateTime", strings.TrimSuffix(string(
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
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := data.Value
			return setInteger(&value)
		}
	case *wrapperspb.UInt64Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			if math.MaxInt64 < data.Value {
				return fmt.Errorf(
					"convert: uint64 %d value into DateTime ; too big ",
					data.Value,
				)
			}
			value := int64(data.Value)
			return setInteger(&value)
		}
	case *wrapperspb.Int32Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := int64(data.Value)
			return setInteger(&value)
		}
	case *wrapperspb.UInt32Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := int64(data.Value)
			return setInteger(&value)
		}
	case *wrapperspb.FloatValue:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := float64(data.Value)
			return setDouble(&value)
		}
	case *wrapperspb.DoubleValue:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := data.Value
			return setDouble(&value)
		}
	case *wrapperspb.StringValue:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := data.Value
			return setString(&value)
		}
	default:
		{
			return fmt.Errorf(
				"convert: %[1]T %[1]v value into DateTime",
				src,
			)
		}
	}
}

func (dv *DateTimeValue) Encode(dst any) error {
	panic("not implemented")
}

// Err to check underlying value
// according to the data type constraints
func (dv *DateTimeValue) Err() error {
	return nil
	panic("not implemented")
}

func (dv *DateTimeValue) Custom(pragma.DoNotImplement) {}

// Accept [pres]:
// - time.Second
// - time.Millisecond
// - time.Microsecond
// - time.Nanosecond
func CastNumberAsDateTime(v float64, pres time.Duration) time.Time {
	const second int64 = 1e9 // time.Second
	toNsec := second / int64(pres)
	// tsec, nsec := math.Modf(v)
	// // return time.Unix(int64(tsec), int64(nsec*float64(pres))*toNsec)
	// return time.Unix(int64(tsec), int64(nsec*float64(toNsec))*int64(pres))
	tsec := int64(v)
	nsec := int64((v * float64(toNsec))) % toNsec
	return time.Unix(tsec, (nsec * int64(pres))).UTC()
	// round := func(num float64) int {
	// 	return int(num + math.Copysign(0.5, num))
	// }
	// toFixed := func(num float64, precision int) float64 {
	// 	output := math.Pow(10, float64(precision))
	// 	return float64(round(num*output)) / output
	// }

	// return time.Unix(int64(tsec), int64(nsec*1e9))
	// tsec, nsec := math.Modf(toFixed(v, 6))
	// nsec = (nsec * precWant) / precWant
	// return time.Unix(int64(tsec), int64((nsec*1e9)/float64(precSkrew)))
	// return time.Unix(int64(tsec), int64(nsec*precWant)*precSkrew)
}

// Cast time.Time value as float64.("tsec[.nsec]") decimal number
func CastDateTimeAsNumber(v time.Time) float64 {
	tsec, nsec := v.Unix(), int64(v.Nanosecond())
	return float64(tsec) + (float64(nsec) / 1e9) // time.Second
}
