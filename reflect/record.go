package customrel

import "github.com/webitel/custom/internal/pragma"

// Record data of the composite type structure
type Record interface {
	// Type of the record structure
	Dataset() DatasetDescriptor

	// Get field value
	Get(FieldDescriptor) any
	// Set field value
	Set(FieldDescriptor, any) error

	// Fields been updated
	Fields() []string
	// Range thru populated field values
	Range(func(fd FieldDescriptor, v any) bool)

	pragma.DoNotImplement
}
