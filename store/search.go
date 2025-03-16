package store

import "context"

type SearchOptions struct {
	// Context
	context.Context
	// Authentication
	Dc, Uid int64
	// Output
	Page   int
	Size   int
	Sort   []string
	Fields []string
	// Request
	Filter map[string]any
}

type SearchOption func(req *SearchOptions)

// NewSearch builds new search request options
func NewSearch(opts ...SearchOption) SearchOptions {
	req := SearchOptions{
		Context: context.Background(),
		Filter:  make(map[string]any),
	}
	for _, setup := range opts {
		setup(&req)
	}
	return req
}

const (
	// Default size of the result page
	DefaultSearchSize = 16 // item(s)
)

func (req *SearchOptions) GetSize() int {
	if req == nil {
		return DefaultSearchSize
	}
	switch {
	case req.Size < 0:
		return -1
	case req.Size > 0:
		// CHECK for too big values !
		return req.Size
	case req.Size == 0:
		return DefaultSearchSize
	}
	panic("unreachable code")
}

func (req *SearchOptions) GetPage() int {
	if req != nil {
		// Limited ? either: manual -or- default !
		if req.GetSize() > 0 {
			// Valid ?page= specified ?
			if req.Page > 0 {
				return req.Page
			}
			// default: always the first one !
			return 1
		}
	}
	// <nop> -or- <nolimit>
	return 0
}
