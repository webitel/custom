package customrel

import (
	"github.com/webitel/custom/internal/pragma"
)

// Type of the primitive data
type Type interface {
	// Kind of the data type.
	Kind() Kind
	// New data value codec.
	New() Codec
	// Err to check data type descriptor integrity.
	Err() error

	// Descriptor() proto.Message

	pragma.DoNotImplement
}
