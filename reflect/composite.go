package customrel

import (
	"github.com/webitel/custom/internal/pragma"
	custompb "github.com/webitel/proto/gen/custom"
)

// DatasetDescriptor represents base composite type structure
type DatasetDescriptor interface {
	// Domain ID that this dataset belongs to
	Dc() int64

	// Err to check the integrity
	// of the data type structure.
	Err() error

	// Name of the dataset type.
	// e.g.: "contacts", "cities".
	// [MUST]. Unique within domain.
	Name() string
	// Relative path to the dataset, aka "package".
	// e.g.: "contacts", "dictionaries/cities".
	// [MUST]. Unique within domain.
	Path() string

	// Fields structure.
	Fields() FieldDescriptors
	// Primary data field.
	Primary() FieldDescriptor
	// Display data field.
	Display() FieldDescriptor
	// Dataset indexing
	Indices() IndexDescriptors

	// ProtoDescriptor as internal specification.
	ProtoDescriptor() *custompb.Dataset

	pragma.DoNotImplement
}

type DictionaryDescriptor interface {
	// implements
	DatasetDescriptor
	// Title of the dataset.
	Title() string
	// Short description of data usage.
	Usage() string

	// Indicates whether this is [ GLOBAL ] dataset type ( .Dc() == 0 ).
	// False - means that this is [ CUSTOM ] type, which data structure MAY be changed.
	IsReadonly() bool
	// Indicates whether this [ GLOBAL ] dataset type supports custom fields extension.
	IsExtendable() bool
}

type ExtensionDescriptor interface {
	// implements
	DatasetDescriptor
	// Dictionary base type [ GLOBAL( readonly ) & extendable ] of the extension.
	Dictionary() DictionaryDescriptor
	// ProtoMessage returns same structure as [ProtoDescriptor] method
	// except that [primary] field definition is hidden for view.
	ProtoMessage() *custompb.Dataset
}

// type Value interface {
// 	Err() error
// 	Interface() any
// }

// type Nullable interface {
// 	IsNull() bool
// }

// func IsNull(v any) bool {
// 	if v == nil {
// 		return true
// 	}
// 	rv
// }

// type DataType interface {
// 	Err() error
// 	Kind() int
// 	New() Value
// }

type FieldDescriptor interface {
	//
	Num() int
	Name() string
	Title() string
	Usage() string

	Kind() Kind
	Type() Type // DataType
	Default() any

	IsPrimary() bool
	IsDisplay() bool

	IsReadonly() bool
	IsRequired() bool
	IsDisabled() bool
	IsHidden() bool

	// Dataset that this Field belongs to ..
	Dataset() DatasetDescriptor
	Descriptor() *custompb.Field
}

type FieldDescriptors interface {
	Num() int
	Get(i int) FieldDescriptor
	ByName(name string) FieldDescriptor
	Range(func(FieldDescriptor) bool)
}

type IndexDescriptor interface {
	//
	Name() string
	IsUnique() bool

	Fields() []string
	Include() []string

	Descriptor() *custompb.Index
}

type IndexDescriptors interface {
	Num() int
	Get(i int) IndexDescriptor
	ByName(name string) IndexDescriptor
	Range(func(IndexDescriptor) bool)
}
