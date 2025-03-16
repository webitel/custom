package data

import (
	"context"
	"strings"

	customrel "github.com/webitel/custom/reflect"
	custompb "github.com/webitel/proto/gen/custom"
	"google.golang.org/protobuf/proto"
)

// ------------ IMPL ----------------- //

// FieldDescriptor of the Dataset structure.
type Field struct {
	ver  uint32
	num  int // positive (non-zero) integer, index(+1) within its collection
	list *Fields
	spec *custompb.Field
	rtyp Type // resolved primitive type
}

var _ customrel.FieldDescriptor = (*Field)(nil)

func (fd *Field) Num() int {
	return fd.num
}

func (fd *Field) Name() string {
	return fd.spec.GetId()
}

// field type kind
// resolve: $(.Type)
func (fd *Field) kindOfType() (kind customrel.Kind) {
	// Field data type descriptor
	spec := fd.spec.GetType()
	if spec == nil {
		// UNDEFINED
		return // customrel.NONE
	}
	// try to resolve by oneof (.Type) declaration
	switch spec.(type) {
	case *custompb.Field_Bool:
		kind = customrel.BOOL
	case *custompb.Field_Int32:
		kind = customrel.INT32
	case *custompb.Field_Int64:
		kind = customrel.INT64
	case *custompb.Field_Int:
		kind = customrel.INT
	case *custompb.Field_Uint32:
		kind = customrel.UINT32
	case *custompb.Field_Uint64:
		kind = customrel.UINT64
	case *custompb.Field_Uint:
		kind = customrel.UINT
	case *custompb.Field_Float32:
		kind = customrel.FLOAT32
	case *custompb.Field_Float64:
		kind = customrel.FLOAT64
	case *custompb.Field_Float:
		kind = customrel.FLOAT
	case *custompb.Field_Binary:
		kind = customrel.BINARY
	case *custompb.Field_Lookup:
		kind = customrel.LOOKUP
	case *custompb.Field_String_:
		kind = customrel.STRING
	case *custompb.Field_Richtext:
		kind = customrel.RICHTEXT
	case *custompb.Field_Datetime:
		kind = customrel.DATETIME
	case *custompb.Field_Duration:
		kind = customrel.DURATION
	}
	return // kind ? customrel.NONE
}

// field data kind
// resolve: $(.Kind) | $(.Type)
func (fd *Field) kindOf() (kind customrel.Kind) {
	spec := fd.spec // Field data type descriptor
	kind = spec.GetKind()
	if kind == customrel.NONE {
		kind = fd.kindOfType()
		spec.Kind = kind // remember
		// [FIXME]: what if undefined here ?
		// will try to detect on every call !
	}
	return // kind
}

// field data type constructor
func (fd *Field) typeOf(kind customrel.Kind) (rtyp Type) {
	// *custompb.Field descriptor
	spec := fd.spec
	switch kind {
	case customrel.LIST:
		// element type
		rtyp = fd.typeOf(
			fd.kindOfType(),
		)
		// list of type
		rtyp = ListAs(rtyp)
	case customrel.BOOL:
		rtyp = BoolAs(spec.GetBool())
	case customrel.INT:
		rtyp = Int.As(spec.GetInt())
	case customrel.INT32:
		rtyp = Int32.As(spec.GetInt32())
	case customrel.INT64:
		rtyp = Int64.As(spec.GetInt64())
	case customrel.UINT:
		rtyp = UnsignedAs(spec.GetUint())
	case customrel.UINT32:
		rtyp = UnsignedAs(spec.GetUint32())
	case customrel.UINT64:
		rtyp = UnsignedAs(spec.GetUint64())
	case customrel.FLOAT:
		rtyp = FloatAs(spec.GetFloat())
	case customrel.FLOAT32:
		rtyp = FloatAs(spec.GetFloat32())
	case customrel.FLOAT64:
		rtyp = FloatAs(spec.GetFloat64())
	case customrel.BINARY:
		rtyp = BinaryAs(spec.GetBinary())
	case customrel.LOOKUP:
		{
			ds := fd.list.typo
			rtyp = LookupAs(nil, ds.Dc(), spec.GetLookup(),
				func(_ context.Context, _ int64, pkg string) (customrel.DictionaryDescriptor, error) {
					if self, ok := ds.(customrel.DictionaryDescriptor); ok {
						eq := strings.EqualFold
						for _, dn := range []string{
							self.Path(), // self.Name(),
						} {
							if eq(dn, pkg) {
								return self, nil
							}
						}
					}
					// Not Found !
					return nil, nil
				})
		}
	case customrel.STRING:
		rtyp = StringAs(spec.GetString_())
	case customrel.RICHTEXT:
		rtyp = StringAs(spec.GetRichtext())
	case customrel.DATETIME:
		rtyp = DateTimeAs(spec.GetDatetime())
	case customrel.DURATION:
		rtyp = DurationAs(spec.GetDuration())
	case customrel.NONE:
		// neither [kind] nor [type] spec
		rtyp = UndefinedAs(ErrNoType)
	default:
		// unknown [kind] spec
		rtyp = UndefinedAs(ErrNoType)
	}
	return // rtyp
}

// Kind of the field data.
func (fd *Field) Kind() customrel.Kind {
	return fd.kindOf()
}

// Type of the field data.
func (fd *Field) Type() customrel.Type {
	if fd.rtyp == nil {
		// resolve: once !
		fd.rtyp = fd.typeOf(
			fd.kindOf(),
		)
	}
	// resolve[d]; once
	return fd.rtyp
}

func (fd *Field) Title() string {
	return fd.spec.GetName()
}

func (fd *Field) Usage() string {
	return fd.spec.GetHint()
}

func (fd *Field) Default() any {
	panic("not implemented") // TODO: Implement
}

func (fd *Field) IsPrimary() bool {
	if ds := fd.list.typo; ds != nil {
		return fd == ds.Primary()
	}
	return false
}

func (fd *Field) IsDisplay() bool {
	if ds := fd.list.typo; ds != nil {
		return fd == ds.Display()
	}
	return false
}

func (fd *Field) IsReadonly() bool {
	return fd.spec.GetReadonly()
}

func (fd *Field) IsRequired() bool {
	return fd.spec.GetRequired()
}

func (fd *Field) IsDisabled() bool {
	return fd.spec.GetDisabled()
}

func (fd *Field) IsHidden() bool {
	return fd.spec.GetHidden()
}

func (fd *Field) Dataset() customrel.DatasetDescriptor {
	return fd.list.typo
}

func (fd *Field) Descriptor() *custompb.Field {
	if fd.spec != nil {
		return proto.Clone(fd.spec).(*custompb.Field)
	}
	return nil
}

// FieldDescriptors collection. Readonly
type Fields struct {
	// ref  *sync.RWMutex
	ver  uint32 // version of the data source
	typo customrel.DatasetDescriptor
	data []*custompb.Field // source
	list []Field           // resolved cache
	hash map[string]*Field // index[name]
}

// type fieldsByName map[string]*Field

// func (c fieldsByName) get(dn string) (fd *Field, ok bool) {
// 	dn = strings.ToLower(dn)
// 	fd, ok = c[dn]
// 	return // fd, ok
// }

// func (c fieldsByName) add(fd *Field) {
// 	dn := strings.ToLower(fd.Name())
// 	c[dn] = fd
// }

// func (c fieldsByName) del(dn string) {
// 	dn = strings.ToLower(dn)
// 	delete(c, dn)
// }

var _ customrel.FieldDescriptors = (*Fields)(nil)

func newFieldDescriptors(ver uint32, typo customrel.DatasetDescriptor, data []*custompb.Field) Fields {
	n := len(data)
	return Fields{
		// ref:  nil,
		ver:  (ver + 1), // MUST: positive !
		typo: typo,
		data: data,
		list: make([]Field, n),
		hash: make(map[string]*Field, n),
	}
}

// Num returns count of the Fields.
func (fs *Fields) Num() int {
	return len(fs.data)
}

// get field by index (offset) number
func (fs *Fields) get(i int) *Field {
	if 0 <= i && i < len(fs.data) {
		fd := &fs.list[i]
		// if fd.desc == nil {
		if fd.ver != fs.ver {
			// DROP ; obsolete
			dn := strings.ToLower(fd.Name())
			delete(fs.hash, dn)
			(*fd) = Field{
				ver:  fs.ver, // ver of the view, created from state !
				num:  (i + 1),
				spec: fs.data[i],
				list: fs,
				rtyp: nil, // NOT resolved yet !
			}
			// [RE]INDEX hash [byName]
			fs.hash[dn] = fd
		}
		return fd
	}
	// Not Found
	return nil
}

// find field by name
func (fs *Fields) find(name string) *Field {
	var (
		dn = strings.ToLower(name) // distinguished name
		eq = strings.EqualFold     // equality matching rule
	)
	fd, ok := fs.hash[dn]
	if fd == nil || fd.ver != fs.ver {
		fd = nil // invalidate !
		fs.Range(func(fx customrel.FieldDescriptor) bool {
			if eq(name, fx.Name()) {
				fd = fx.(*Field)
				return false // break ; Found !
			}
			return true // continue ; Not Found !
		})
		if fd != nil {
			// Found !
			fs.hash[dn] = fd
		} else if ok {
			// Invalidate !
			delete(fs.hash, dn)
		}
	}
	return fd // nil?
}

// Get FieldDescriptor by index / position
func (fs *Fields) Get(i int) customrel.FieldDescriptor {
	fd := fs.get(i)
	if fd != nil {
		return fd
	}
	return nil
}

// Get FieldDescriptor by it's name.
func (fs *Fields) ByName(name string) customrel.FieldDescriptor {
	fd := fs.find(name)
	if fd != nil {
		return fd
	}
	return nil
}

// Range iterates over all collection
func (fs *Fields) Range(next func(customrel.FieldDescriptor) bool) {
	count := fs.Num()
	for i := range count {
		if !next(fs.Get(i)) {
			break // return
		}
	}
}
