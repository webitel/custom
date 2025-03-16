package customrel

import datapb "github.com/webitel/proto/gen/custom/data"

type Kind = datapb.Kind

const (
	NONE     Kind = datapb.Kind_none
	LIST          = datapb.Kind_list
	BOOL          = datapb.Kind_bool
	INT           = datapb.Kind_int
	INT32         = datapb.Kind_int32
	INT64         = datapb.Kind_int64
	UINT          = datapb.Kind_uint
	UINT32        = datapb.Kind_uint32
	UINT64        = datapb.Kind_uint64
	FLOAT         = datapb.Kind_float
	FLOAT32       = datapb.Kind_float32
	FLOAT64       = datapb.Kind_float64
	BINARY        = datapb.Kind_binary
	LOOKUP        = datapb.Kind_lookup
	STRING        = datapb.Kind_string
	RICHTEXT      = datapb.Kind_richtext
	DATETIME      = datapb.Kind_datetime
	DURATION      = datapb.Kind_duration
)
