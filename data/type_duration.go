package data

import datapb "github.com/webitel/proto/gen/custom/data"

type Duration struct {
	spec *datapb.Duration
}

func DurationAs(spec *datapb.Duration) Type {
	panic("not implemented")
}
