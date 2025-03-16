package data

import (
	datapb "github.com/webitel/proto/gen/custom/data"
)

type Binary struct {
	spec *datapb.Binary
}

func BinaryAs(spec *datapb.Binary) Type {
	panic("not implemented")
}
