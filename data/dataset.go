package data

import (
	"errors"
	"fmt"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/proto"
)

// dataset base composite type descriptor
type dataset struct {
	dc     int64
	err    error // once: integrity error
	spec   *custompb.Dataset
	fields Fields
}

func datasetAs(dc int64, spec *custompb.Dataset) *dataset {
	if dc < 0 {
		dc = 0 // normalized
	}
	ds := &dataset{
		dc: dc, spec: spec,
	}
	// init fields
	ds.fields = newFieldDescriptors(
		0, ds, ds.spec.Fields,
	)
	return ds
}

var _ customrel.DatasetDescriptor = (*dataset)(nil)

// Domain ID that this dataset belongs to
func (ds *dataset) Dc() int64 {
	return ds.dc
}

// Err to check the integrity
// of the data type structure.
func (ds *dataset) Err() error {
	if ds == nil {
		// Not Found
	}
	if ds.err != nil {
		// once: critical
		return ds.err
	}
	var (
		errs, err error
		withErr   = func(err error) {
			errs = errors.Join(errs, err)
		}
	)
	defer func() {
		if errs != nil {
			// once: failed !
			ds.err = err
		}
	}()
	if ds.spec == nil {
		// Not Found
	}
	// [TODO]: Validate !
	// [CHECK]:
	// - name
	// - path
	// - fields
	// - indices
	// - [primary] ?
	// - [display] ?

	// [CHECK] fields
	fields := ds.Fields()
	fields.Range(func(fd customrel.FieldDescriptor) bool {
		name := fd.Name() // ^\w+$
		// IS UNIQUE [name] ?!
		if num := fd.Num(); num > 1 {
			// ordered: [ 1, 2, 3, .. ]
			fields.Range(func(fx customrel.FieldDescriptor) bool {
				// if CaseIgnoreMatch(fd.GetId(), e.GetId()) {
				if name == fx.Name() {
					withErr(fmt.Errorf("fields( name: %s ); duplicate", name))
					return false
				}
				// check previous fields only !
				return (fx.Num() + 1) < num
			})
		}
		// field.kind
		// field.type
		if err = fd.Type().Err(); err != nil {
			withErr(fmt.Errorf("fields( name: %s ); %v", name, err))
		}
		// field.value; default
		return true // (err == nil)
	})
	return errs
}

// Name of the dataset type.
// e.g.: "contacts", "cities".
// [MUST]. Unique within domain.
func (ds *dataset) Name() string {
	return ds.spec.GetRepo()
}

// Relative path to the dataset, aka "package".
// e.g.: "contacts", "dictionaries/cities".
// [MUST]. Unique within domain.
func (ds *dataset) Path() string {
	return ds.spec.GetPath()
}

// Fields structure.
func (ds *dataset) Fields() customrel.FieldDescriptors {
	// if ds != nil {
	return &ds.fields
	// }
	// return nil
}

// Primary data field.
func (ds *dataset) Primary() customrel.FieldDescriptor {
	if ds != nil {
		return ds.Fields().ByName(ds.spec.Primary)
	}
	return nil
}

// Display data field.
func (ds *dataset) Display() customrel.FieldDescriptor {
	if ds != nil {
		return ds.Fields().ByName(ds.spec.Display)
	}
	return nil
}

// Dataset indexing
func (ds *dataset) Indices() customrel.IndexDescriptors {
	panic("not implemented") // TODO: Implement
}

// IsReadonly reports whether this is [ GLOBAL ] type descriptor.
func (ds *dataset) IsReadonly() bool {
	// Has INVALID domain ID assigned !
	return ds.dc < 1
}

func (ds *dataset) ProtoDescriptor() *custompb.Dataset {
	if ds.spec != nil {
		return proto.Clone(ds.spec).(*custompb.Dataset)
	}
	return nil
}

// pragma.DoNotImplement
func (*dataset) Custom(pragma.DoNotImplement) {}
