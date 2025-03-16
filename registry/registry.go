package customreg

import (
	"context"
	"slices"
	"strings"
	"sync"

	customrel "github.com/webitel/custom/reflect"
)

// type aliases
type (
	Dataset    = customrel.DatasetDescriptor
	Extension  = customrel.ExtensionDescriptor
	Dictionary = customrel.DictionaryDescriptor

	domain = cache[Dataset]
)

var (
	indexKey  = strings.ToLower
	indexKeys = func(regtype Dataset) []any {
		// return []any{ // [MUST]: UNIQUE //
		// 	indexKey(regtype.Name()),
		// 	indexKey(regtype.Path()),
		// }
		keys := []any{ // [MUST]: UNIQUE //
			indexKey(regtype.Name()),
			indexKey(regtype.Path()),
		}
		n := len(keys)
		remove := func(e int) {
			// keys = append(keys[:a], keys[a+1:]...)
			keys = slices.Delete(keys, e, e+1)
			n--
		}
		for a := 0; a < n; a++ {
			if keys[a] == "" {
				// invalid !
				remove(a)
				a--
				continue
			}
			for b := (a - 1); b >= 0; b-- {
				if keys[a] == keys[b] {
					// duplicate !
					remove(a)
					a--
					break
				}
			}
		}
		return keys
	}
)

func newDomain(size int) *domain {
	return newCache(
		indexKeys, size,
	)
}

// GLOBAL registry
type registry struct {
	mu sync.Mutex
	dc map[int64]*domain
}

// KEEP LOCKED !
func (c *registry) lazyInit() {
	if c.dc == nil {
		c.dc = make(map[int64]*domain)
		c.dc[0] = newDomain(1 << 8) // GLOBAL
	}
}

func (c *registry) domain(dc int64) *domain {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dc[dc]
}

func (c *registry) Lookup(dc int64, typeOf string) Dataset {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyInit()

	if domain := c.dc[dc]; domain != nil {
		return domain.Get(typeOf)
	}
	return nil
}

func (c *registry) Register(ds Dataset) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyInit()

	var (
		err error
		pdc = ds.Dc()
		dc  = c.dc[pdc]
	)

	if dc == nil {
		dc = newDomain(1 << 8)
		defer func() {
			if err == nil {
				// NEW domain !
				c.dc[pdc] = dc
			}
		}()
	}

	err = dc.Add(ds)
	return err
}

func (c *registry) Unregister(ds Dataset) error {
	// panic("not implemented")

	pdc := ds.Dc()
	if pdc < 1 {
		// Unregister [GLOBAL] types DISALLOWED !
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lazyInit()

	dc := c.dc[pdc]
	if dc == nil {
		// Not Found
		return nil
	}

	ok := dc.Del(ds)
	if ok && dc.Num() == 0 {
		delete(c.dc, pdc)
	}
	return nil
}

type Types struct {
	registry *registry
	resolver CustomTypeResolver
	// mu    sync.Mutex // protects [resolver] interface calls !
}

// func newTypes() *Types {
// 	return &Types{
// 		// domains: make(map[int64]*domainDescriptors),
// 		registry: &registry{},
// 	}
// }

// func (c *Types) init() {
// 	// if c.domains == nil {
// 	// 	c.domains = make(map[int64]*domainDescriptors)
// 	// }
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
// 	if c.registry == nil {
// 		c.registry = &registry{}
// 	}
// }

func (c *Types) Register(ds Dataset) error {
	// c.init()
	return c.registry.Register(ds)
}

func (c *Types) Unregister(ds Dataset) error {
	// c.init()
	return c.registry.Unregister(ds)
}

func (c *Types) WithResolver(impl CustomTypeResolver) *Types {
	// c.init()
	if is, _ := impl.(*Types); is == c {
		return c // SELF
	}

	if c.resolver != nil {
		is := c.resolver.(*protectedTypeResolver)
		if is.CustomTypeResolver == impl {
			return c // SAME
		}
	}

	return &Types{
		registry: c.registry, // CHAIN
		resolver: &protectedTypeResolver{
			CustomTypeResolver: impl,
		},
	}
}

func (c *Types) GetExtension(ctx context.Context, dc int64, pkg string) (typ Extension, err error) {
	// panic("not implemented")

	if dc < 1 {
		// Invalid [d]omain[c]omponent id spec.
		// Not Found !
		return // nil, nil
	}
	// normalized index value
	pkg = indexKey(pkg)
	if pkg == "" {
		// Missing [SUPER] type spec.
		// Not Found !
		return // nil, nil
	}

	// c.init()
	if reg := c.registry.Lookup(dc, pkg); reg != nil {
		if typ, _ = reg.(Extension); typ != nil {
			return typ, nil // [FROM]: Cache Found !
		}
	}

	// [CUSTOM] resolution !

	defer func() {
		// Add to cache -if- resolved !
		if typ != nil && err == nil {
			_ = c.registry.Register(typ)
		}
	}()

	// c.mu.Lock()
	// defer c.mu.Unlock()

	if dc > 0 && c.resolver != nil {
		typ, err = c.resolver.GetExtension(
			ctx, dc, pkg,
		)
		if err != nil {
			typ = nil
			return // nil, err
		}
		if typ != nil && typ.Dc() != dc {
			// Invalid result !
			typ = nil
		}
		return // typ?, nil
	}

	// Not Found !
	return // nil, nil
}

func (c *Types) GetDictionary(ctx context.Context, dc int64, pkg string) (typ Dictionary, err error) {
	// panic("not implemented")

	if dc < 0 {
		dc = 0 // GLOBAL
	}
	// normalized index value
	pkg = indexKey(pkg)
	if pkg == "" {
		// Missing [DATASET] type spec.
		// Not Found !
		return // nil, nil
	}

	// c.init()
	for _, pdc := range []int64{dc, 0} {
		reg := c.registry.Lookup(pdc, pkg)
		if typ, _ = reg.(Dictionary); typ != nil {
			return // typ, nil // [FROM]: Cache Found !
		}
		if pdc == 0 {
			break
		}
	}

	// [CUSTOM] resolution !

	defer func() {
		// Add to cache -if- resolved NON [GLOBAL] !
		if err == nil && typ != nil && typ.Dc() > 0 {
			_ = c.registry.Register(typ)
		}
	}()

	// c.mu.Lock()
	// defer c.mu.Unlock()

	if dc > 0 && c.resolver != nil {
		typ, err = c.resolver.GetDictionary(
			ctx, dc, pkg,
		)
		if err != nil {
			// typ = nil
			return // nil, err
		}
		if typ != nil && typ.Dc() != dc {
			if typ.Dc() != 0 {
				// Invalid result
				typ = nil
			}
		}
		return // typ?, nil
	}
	// Not Found !
	return // nil, nil
}

// RangeDictionaries of [GLOBAL] domain ( dc == 0 ) ONLY !
func (c *Types) RangeDictionaries(next func(reg Dictionary) bool) {
	// c.init()
	global := c.registry.domain(0)
	global.Range(func(reg Dataset) bool {
		e, _ := reg.(Dictionary)
		if e == nil {
			// Not Dictionary ; Next ..
			return true
		}
		return next(e)
	})
}

// GlobalTypes REGISTRY cache
var GlobalTypes = &Types{
	registry: &registry{},
	// domains: make(map[int64]*domainDescriptors),
}

func Register(ds Dataset) error {
	return GlobalTypes.Register(ds)
}

func Unregister(ds Dataset) error {
	return GlobalTypes.Unregister(ds)
}

// Invalidate removes CUSTOM dataset record from cache.
func Invalidate(dc int64, pkg string) error {
	regtyp := GlobalTypes.registry.Lookup(dc, pkg)
	if regtyp != nil {
		return GlobalTypes.Unregister(regtyp)
	}
	// Not Found !
	return nil
}

// GetExtension is shorthand of GlobalTypes.GetExtension(!)
func GetExtension(ctx context.Context, dc int64, pkg string) (Extension, error) {
	return GlobalTypes.GetExtension(ctx, dc, pkg)
}

// GetExtension is shorthand of GlobalTypes.GetDictionary(!)
func GetDictionary(ctx context.Context, dc int64, pkg string) (Dictionary, error) {
	return GlobalTypes.GetDictionary(ctx, dc, pkg)
}
