package data

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	datapb "github.com/webitel/proto/gen/custom/data"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Duration struct {
	spec *datapb.Duration
}

func DurationAs(spec *datapb.Duration) Type {
	if spec == nil {
		return (*Duration)(nil)
	}
	return &Duration{spec: spec}
}

var _ Type = (*Duration)(nil)

// Kind of the data type.
func (*Duration) Kind() customrel.Kind {
	return customrel.DURATION
}

// New data value codec.
func (dt *Duration) New() customrel.Codec {
	return &DurationValue{typof: dt}
}

// Err to check data type descriptor integrity.
func (dt *Duration) Err() error {
	// [TODO]
	return nil
	panic("not implemented")
}

func (*Duration) Custom(pragma.DoNotImplement) {}

var durationError = map[string]string{
	"min": "duration value {{.value}} violates min boundary of {{.min}}",
	"max": "duration value {{.value}} violates max boundary of {{.max}}",
}

func (dt *Duration) violationError(kind string, val *time.Duration) error {
	tmpl := durationError[kind]
	tmpl = strings.ReplaceAll(tmpl, "{{.value}}", fmt.Sprintf("%v", val))
	tmpl = strings.ReplaceAll(tmpl, "{{.min}}", fmt.Sprintf("%v", time.Duration(dt.spec.Min.GetValue()*int64(time.Second))))
	tmpl = strings.ReplaceAll(tmpl, "{{.max}}", fmt.Sprintf("%v", time.Duration(dt.spec.Max.GetValue()*int64(time.Second))))
	return RequestError(
		fmt.Sprintf("custom.type.duration.%s.violation", kind), tmpl,
	)
}

// Accept checks value type constraints ..
func (dt *Duration) Accept(val *time.Duration) error {
	// if val == nil {
	// 	return nil // NULL ; [OK]
	// }
	if dt == nil || dt.spec == nil {
		// no type constraints assigned !
		return nil // [OK] ; whatever ..
	}
	var (
		min = dt.spec.Min
		max = dt.spec.Max
		// notnull = (min != nil)
	)
	if val == nil {
		if min == nil {
			return nil // [OK]
		}
		// [NOTNULL]
		return dt.violationError("min", val)
	}
	sec := int64(*val / time.Second)
	if min != nil && sec < min.Value {
		return dt.violationError("min", val)
	}
	if max != nil && max.Value < sec {
		return dt.violationError("max", val)
	}
	return nil // [OK]
}

type DurationValue struct {
	typof *Duration
	value *time.Duration
}

var _ customrel.Codec = (*DurationValue)(nil)

// Type implements customrel.Codec.
func (dv *DurationValue) Type() customrel.Type {
	if dv != nil {
		return dv.typof
	}
	return (*Duration)(nil)
}

// Custom implements customrel.Codec.
func (*DurationValue) Custom(pragma.DoNotImplement) {}

// Decode implements customrel.Codec.
func (dv *DurationValue) Decode(src any) error {
	setValue := func(set *time.Duration) (err error) {
		if err = dv.typof.Accept(set); err == nil {
			dv.value = set // [OK]
		}
		return // err?
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
	setDouble := func(set *float64) error {
		if set == nil {
			return setValue(nil)
		}
		value := CastNumberAsDuration(*set, time.Millisecond)
		return setValue(&value)
	}
	setSecond := func(set *int64) error {
		if set == nil {
			return setValue(nil)
		}
		value := time.Duration(*set) * time.Second
		return setValue(&value)
	}
	setString := func(set *string) (err error) {
		if set == nil {
			return setValue(nil)
		}
		var (
			dur  time.Duration
			text = strings.TrimSpace(*set)
		)
		switch text {
		case "", "0", "0s":
			return setValue(&dur) // Zero(0)
		}
		// Postgres-style ; REFACTOR !
		var interval pgtype.Interval
		if err = interval.Scan(text); err == nil && interval.Valid {
			// https://github.com/jackc/pgx/blob/v5.7.4/pgtype/builtin_wrappers.go#L507
			const (
				microsecondsPerSecond = 1000000 // time.Second / time.Microsecond
				microsecondsPerMinute = 60 * microsecondsPerSecond
				microsecondsPerHour   = 60 * microsecondsPerMinute
				microsecondsPerDay    = 24 * microsecondsPerHour
				microsecondsPerMonth  = 30 * microsecondsPerDay
			)
			us := int64(interval.Months)*microsecondsPerMonth +
				int64(interval.Days)*microsecondsPerDay +
				interval.Microseconds
			dur = time.Duration(
				time.Duration(us) * time.Microsecond,
			)
			return setValue(&dur) // [OK]
		}
		// GoLang-style
		if dur, err = time.ParseDuration(text); err == nil {
			return setValue(&dur) // [OK]
		}
		err = RequestError(
			"custom.type.duration.cast.error",
			"custom: cannot cast string %s into Duration",
			*set,
		)
		return // err
	}
	switch data := src.(type) {
	case *DurationValue:
		{
			if data == dv {
				return nil // [OK] ; self
			}
			return setValue(data.value)
		}
	case DurationValue:
		{
			return setValue(data.value)
		}
	case int64:
		{
			return setSecond(&data)
		}
	case *int64:
		{
			return setSecond(data)
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
	case *structpb.Value:
		{
			// if data == nil {
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
						"convert: %s value %v into Duration", strings.TrimSuffix(string(
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
			return setSecond(&value)
		}
	case *wrapperspb.UInt64Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			if math.MaxInt64 < data.Value {
				return fmt.Errorf("convert: Uint64 value %d into Duration ; too big", data.Value)
			}
			value := int64(data.Value)
			return setSecond(&value)
		}
	case *wrapperspb.Int32Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := int64(data.Value)
			return setSecond(&value)
		}
	case *wrapperspb.UInt32Value:
		{
			// if data == nil {
			// 	return setValue(nil)
			// }
			value := int64(data.Value)
			return setSecond(&value)
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
			value := float64(data.Value)
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
			return RequestError(
				"custom.type.duration.cast.error",
				"custom: cannot cast %[1]T value %[1]v into Duration",
				src,
			)
		}
	}
	panic("unreachable code")
}

// Encode implements customrel.Codec.
func (dv *DurationValue) Encode(dst any) error {
	panic("unimplemented")
}

// Err implements customrel.Codec.
func (dv *DurationValue) Err() (err error) {
	if dv == nil {
		return // [OK]
	}
	if dv.typof != nil {
		err = dv.typof.Err()
		if err != nil {
			return // [ERR]
		}
	}
	return // [OK]
}

// Interface of the [*time.Duration] value.
func (dv *DurationValue) Interface() any {
	if dv != nil {
		return dv.value //
	}
	return (*time.Duration)(nil)
}

func CastNumberAsDuration(num float64, pres time.Duration) time.Duration {
	const second int64 = 1e9 // time.Second
	toNsec := second / int64(pres)
	// tsec, nsec := math.Modf(v)
	// // return time.Unix(int64(tsec), int64(nsec*float64(pres))*toNsec)
	// return time.Unix(int64(tsec), int64(nsec*float64(toNsec))*int64(pres))
	tsec := int64(num)
	nsec := int64((num * float64(toNsec))) % toNsec
	return time.Duration((tsec * second) + (nsec * int64(pres)))
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

func CastDurationAsNumber(v time.Duration) float64 {
	return v.Seconds()
}
