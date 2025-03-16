package store

import (
	"context"

	"github.com/webitel/custom/data"
	customrel "github.com/webitel/custom/reflect"
	customreg "github.com/webitel/custom/registry"
)

type customTypeResolver struct {
	Catalog
}

var _ customreg.CustomTypeResolver = customTypeResolver{}

func CustomTypeResolver(schema Catalog) customreg.CustomTypeResolver {
	return customTypeResolver{Catalog: schema}
}

// FindDictionary looks up a dataset structure by its relative package path.
// e.g., "dictionaries/cities", "call_center/agents", "contacts".
//
// This return (nil, NotFound) if not found.
func (c customTypeResolver) GetDictionary(ctx context.Context, dc int64, pkg string) (customrel.DictionaryDescriptor, error) {
	// panic("not implemented")
	// page, err := c.Catalog.Search(ctx, dc, "dictionaries", pkg)
	page, err := c.Catalog.Search(func(req *SearchOptions) {

		req.Context = ctx
		req.Dc = dc

		req.Page = 1
		req.Size = 1
		req.Fields = []string{"+"}

		req.Filter["dir"] = "dictionaries"
		req.Filter["path"] = pkg

	})
	if err != nil {
		return nil, err
	}
	if page.GetNext() || len(page.GetData()) != 1 {
		// Not Found !
		return nil, nil
	}
	spec := page.Data[0]
	if spec.GetReadonly() {
		return customreg.GetDictionary(
			ctx, 0, pkg, // [ GLOBAL ]
		)
	}
	return data.DictionaryOf(dc, spec), nil
}

// GetExtension looks up a dataset structure by its relative package path to the parent (extendable) dictionary type.
// e.g., "contacts", "cases"
//
// This return (nil, NotFound) if not found.
func (c customTypeResolver) GetExtension(ctx context.Context, dc int64, pkg string) (customrel.ExtensionDescriptor, error) {
	// panic("not implemented")
	// page, err := c.Catalog.Search(ctx, dc, "extensions", pkg)
	page, err := c.Catalog.Search(func(req *SearchOptions) {

		req.Context = ctx
		req.Dc = dc

		req.Page = 1
		req.Size = 1
		req.Fields = []string{"+"}

		req.Filter["dir"] = "extensions"
		req.Filter["path"] = pkg

	})
	if err != nil {
		return nil, err
	}
	if page.GetNext() || len(page.GetData()) != 1 {
		// Not Found !
		return nil, nil
	}
	spec := page.Data[0]
	return data.ExtensionOf(dc, spec), nil
}
