package data

import (
	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
)

const (
	DictionariesDir = "dictionaries"
)

type Dictionary struct {
	*dataset // embeded
}

// DictionaryOf returns [custom/reflect.DictionaryDescriptor]
// for given [*custompb.Dataset] type specification.
func DictionaryOf(dc int64, spec *custompb.Dataset) Dictionary {
	return Dictionary{datasetAs(dc, spec)}
}

// NewDictionary returns [custom/reflect.DictionaryDescriptor]
// for given [*custompb.InputDictionary] type specification.
func NewDictionary(dc int64, spec *custompb.InputDictionary) (Dictionary, error) {
	panic("not implemented")
	// return Dictionary{datasetOf(dc, spec)}
}

var _ customrel.DictionaryDescriptor = Dictionary{}

// // Err to check the integrity
// // of the data type structure.
// func (dt Dictionary) Err() error

// Title of the dataset.
func (dt Dictionary) Title() string {
	return dt.spec.GetName()
}

// Short description of data usage.
func (dt Dictionary) Usage() string {
	return dt.spec.GetAbout()
}

// Indicates whether this is [GLOBAL] dataset type ( .Dc() == 0 ). Can edit ?
func (dt Dictionary) IsReadonly() bool {
	return dt.Dc() < 1
}

// Indicates whether this [GLOBAL] type supports custom fields extension.
func (dt Dictionary) IsExtendable() bool {
	return dt.IsReadonly() && dt.dataset.spec.GetExtendable()
}
