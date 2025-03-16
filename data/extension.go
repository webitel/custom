package data

import (
	"context"
	"fmt"
	"path"

	"github.com/webitel/custom/internal/pragma"
	customrel "github.com/webitel/custom/reflect"
	customreg "github.com/webitel/custom/registry"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/proto"
)

const (
	ExtensionsDir = "extensions"
)

type Extension struct {
	sub *dataset
	sup customrel.DictionaryDescriptor
	err error // fatal error
}

// ExtensionOf specification descriptor
func ExtensionOf(dc int64, spec *custompb.Dataset) *Extension {
	if dc < 1 {
		panic("custom: extension domain required")
	}
	spec = proto.Clone(spec).(*custompb.Dataset)
	base, err := extensionFor(
		context.TODO(), spec.GetRepo(),
	)
	if err != nil {
		// failed to resolve base type of extension
		return &Extension{
			sub: datasetAs(dc, spec),
			sup: base, // ?
			err: err,  // !
		}
	}
	// WITH [primary] field !
	fields := make([]*custompb.Field, 1, (1 + len(spec.GetFields())))
	// Copy [primary] field FROM [SUPER] type !
	primary := base.Primary().Descriptor()
	// [re]set ; actualize !
	spec.Primary = primary.GetId()
	spec.Display = spec.Primary
	fields[0] = primary
	// Append custom [extension] fields..
	// copy(fields[1:], as.GetFields())
	for _, field := range spec.GetFields() {
		if field.Id == primary.Id {
			continue // already
		}
		fields = append(fields, field)
	}
	spec.Fields = fields
	return &Extension{
		sub: datasetAs(dc, spec),
		sup: base,
	}
}

// NewExtension specification
func NewExtension(dc int64, spec *custompb.InputExtension) (*Extension, error) {
	panic("not implemented")
}

var _ customrel.ExtensionDescriptor = (*Extension)(nil)

// Domain ID that this dataset belongs to
func (xt *Extension) Dc() int64 {
	return xt.sub.Dc()
}

// Err to check the integrity
// of the data type structure.
func (xt *Extension) Err() error {
	if xt == nil {
		// NotFound
	}
	if xt.err != nil {
		// once
		return xt.err
	}
	if xt.sup == nil {
		// No [SUPER] type descriptor !
	}
	if xt.sub == nil {
		// No [CUSTOM] type descriptor !
	}
	if xt.Dc() < 1 {
		return fmt.Errorf("custom: extension domain required")
	}
	if err := xt.sup.Err(); err != nil {
		return err
	}
	if err := xt.sub.Err(); err != nil {
		return err
	}
	// // resolve base dictionary type
	// _, err := xt.baseDictionary(nil)
	// if err != nil {
	// 	// Not Found !
	// 	// Not Extendable !
	// 	return err
	// }
	// Has field(s) assigned !
	return nil
}

// Name of the dataset type.
// e.g.: "contacts", "cities".
// [MUST]. Unique within domain.
func (xt *Extension) Name() string {
	if xt.sup != nil {
		return xt.sup.Name()
	}
	// UNDEFINED
	return ""
}

// Relative path to the dataset, aka "package".
// e.g.: "contacts", "dictionaries/cities".
// [MUST]. Unique within domain.
func (xt *Extension) Path() string {
	return path.Join(
		ExtensionsDir, xt.Name(),
	)
}

// Fields structure.
func (xt *Extension) Fields() customrel.FieldDescriptors {
	return xt.sub.Fields()
}

// Primary data field.
func (xt *Extension) Primary() customrel.FieldDescriptor {
	return xt.Dictionary().Primary()
}

// Display data field.
func (xt *Extension) Display() customrel.FieldDescriptor {
	return xt.Dictionary().Primary()
}

// Dataset indexing
func (xt *Extension) Indices() customrel.IndexDescriptors {
	return xt.sub.Indices()
}

// resolve base Dictionary type of the Extension
func extensionFor(ctx context.Context, pkg string) (base customrel.DictionaryDescriptor, err error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	// resolve & load
	base, err = customreg.GetDictionary(
		ctx, 0, pkg, // [ GLOBAL & extendable ]
	)
	if err != nil {
		base = nil
		return // nil, err
	}
	if base == nil {
		// Not Found !
		err = fmt.Errorf("custom: %s type not found", pkg)
		return // nil, NotFound
	}
	if !base.IsExtendable() {
		// Not Extendable !
		base = nil
		err = fmt.Errorf("custom: %s type not extendable", base.Path())
		return // nil, NotExtendable
	}
	// resolved !
	return // base, nil
}

// resolve base Dictionary type of the Extension
func (xt *Extension) baseDictionary(ctx context.Context) (base customrel.DictionaryDescriptor, err error) {
	// resolved ?
	if xt.sup != nil {
		// found !
		return xt.sup, nil
	}
	if ctx == nil {
		ctx = context.TODO()
	}
	// resolve & load
	base, err = customreg.GetDictionary(
		ctx, 0, xt.sub.Name(),
	)
	if err != nil {
		base = nil
		return // nil, err
	}
	if base == nil {
		// Not Found !
		err = fmt.Errorf("custom: %s base type not found", xt.Path())
		return // nil, NotFound
	}
	if !base.IsExtendable() {
		// Not Extendable !
		base = nil
		err = fmt.Errorf("custom: %s base type not extendable", xt.Path())
		return // nil, NotExtendable
	}
	// resolved !
	xt.sup = base
	return // base, nil
}

// Dictionary base type [ GLOBAL( readonly ) & extendable ] of the extension.
func (xt *Extension) Dictionary() customrel.DictionaryDescriptor {
	base, _ := xt.baseDictionary(nil)
	return base
}

// ExtensionDescriptor. [NOTE]: internal USE only !
// Returns valid [Dataset] specification with [primary] field
// assigned from [SUPER] type for structure data consistency.
func (xt *Extension) ProtoDescriptor() *custompb.Dataset {
	return &custompb.Dataset{
		Repo:       xt.Name(),
		Path:       xt.Path(),
		Name:       "",
		About:      "",
		Fields:     xt.sub.fields.data,
		Primary:    xt.Primary().Name(),
		Display:    xt.Display().Name(),
		Indices:    xt.sub.spec.GetIndices(), // map[string]*custompb.Index{},
		Readonly:   false,
		Extendable: false,
		CreatedAt:  xt.sub.spec.CreatedAt,
		CreatedBy:  xt.sub.spec.CreatedBy,
		UpdatedAt:  xt.sub.spec.UpdatedAt,
		UpdatedBy:  xt.sub.spec.UpdatedBy,
	}
}

// ProtoMessage returns same structure as [Descriptor] method
// except [primary] field(s) declaration as is NOT available for view.
func (xt *Extension) ProtoMessage() *custompb.Dataset {
	base := xt.Dictionary()
	desc := xt.sub
	fields := xt.Fields()
	primary := base.Primary().Name()
	// Build
	view := &custompb.Dataset{
		Repo: xt.Name(),
		Path: xt.Path(),
		// Name:         "",
		// About:        "",
		Fields: make([]*custompb.Field, 0, fields.Num()),
		// Primary:      "", // HIDDEN
		// Display:      "", // HIDDEN
		Indices: desc.spec.GetIndices(), // map[string]*pbdata.Index{},
		// Readonly:     false,
		// Extendable:   false,
		// Administered: false,
		// Objclass:     "",
		CreatedAt: desc.spec.CreatedAt, // 0
		CreatedBy: desc.spec.CreatedBy, // &pbdata.LookupValue{},
		UpdatedAt: desc.spec.UpdatedAt, // 0,
		UpdatedBy: desc.spec.UpdatedBy, // &pbdata.LookupValue{},
	}
	// Exclude [primary] from public view !
	fields.Range(func(fd customrel.FieldDescriptor) bool {
		if fd.Name() == primary {
			return true // HIDDEN, go next
		}
		view.Fields = append(view.Fields, fd.Descriptor())
		return true // next
	})

	return view
}

func (*Extension) Custom(pragma.DoNotImplement) {}
