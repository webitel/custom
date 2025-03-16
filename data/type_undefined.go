package data

import (
	"fmt"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
)

var ErrNoType = fmt.Errorf("custom: [type] undefined ")

type Undefined struct {
	err error
}

func UndefinedAs(err error) Type {
	return Undefined{err: err}
}

var _ customrel.Type = Undefined{}

func (Undefined) Custom(pragma.DoNotImplement) {}

// Kind of the data type.
func (Undefined) Kind() customrel.Kind {
	return customrel.NONE
}

// New data value codec.
func (dt Undefined) New() customrel.Codec {
	return dt
}

// Err to check data type descriptor integrity.
func (dt Undefined) Err() error {
	if dt.err != nil {
		return dt.err
	}
	return ErrNoType
}

var _ customrel.Codec = Undefined{}

// Type of the data value
func (dv Undefined) Type() customrel.Type {
	return dv
}

// Interface of the underlying value.
// *int[32|64]
// *uint[32|64]
// *float[32|64]
// *time.Time
// *time.Duration
// *Lookup
// []any | LIST
// map[string]any | RECORD
func (dv Undefined) Interface() any {
	return nil // dv.Err()
}

func (dv Undefined) Decode(_ any) error {
	return dv.Err()
}

func (dv Undefined) Encode(_ any) error {
	return dv.Err()
}
