package store

import (
	custompb "github.com/webitel/proto/gen/custom"
)

type Catalog interface {
	Search(opts ...SearchOption) (*custompb.DatasetList, error)
}
