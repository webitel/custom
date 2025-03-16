package customreg

import (
	"context"
	"sync"
)

type DictionaryTypeResolver interface {
	// FindDictionary looks up a dataset structure by its relative package path.
	// e.g., "dictionaries/cities", "call_center/agents", "contacts".
	//
	// This return (nil, NotFound) if not found.
	GetDictionary(ctx context.Context, dc int64, pkg string) (Dictionary, error)
}

type ExtensionTypeResolver interface {
	// GetExtension looks up a dataset structure by its relative package path to the parent (extendable) dictionary type.
	// e.g., "contacts", "cases"
	//
	// This return (nil, NotFound) if not found.
	GetExtension(ctx context.Context, dc int64, pkg string) (Extension, error)
}

type CustomTypeResolver interface {
	DictionaryTypeResolver
	ExtensionTypeResolver
}

type protectedTypeResolver struct {
	CustomTypeResolver
	extensionsLoad   sync.Mutex
	dictionariesLoad sync.Mutex
}

var _ CustomTypeResolver = (*protectedTypeResolver)(nil)

// GetDictionary looks up a struct by its full [pkg] path.
// E.g., "dictionaries/cities", "call_center/agents", "roles"
//
// This return (nil, NotFound) if not found.
func (c *protectedTypeResolver) GetDictionary(ctx context.Context, dc int64, typeOf string) (Dictionary, error) {
	c.dictionariesLoad.Lock()
	defer c.dictionariesLoad.Unlock()
	return c.CustomTypeResolver.GetDictionary(ctx, dc, typeOf)
}

// GetExtension looks up a struct type by its full [pkg] path of the parent (extendable) type.
// E.g., "contacts", "cases"
//
// This return (nil, NotFound) if not found.
func (c *protectedTypeResolver) GetExtension(ctx context.Context, dc int64, typeOf string) (Extension, error) {
	c.extensionsLoad.Lock()
	defer c.extensionsLoad.Unlock()
	return c.CustomTypeResolver.GetExtension(ctx, dc, typeOf)
}
