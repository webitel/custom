package postgres

import (
	"database/sql"

	"github.com/webitel/custom/store"
)

type (
	dataScanFunc[TRow any] func(TRow) sql.Scanner
	dataScanPlan[TRow any] []dataScanFunc[TRow]
)

type query[TRow any] struct {
	req  store.SearchOptions
	plan dataScanPlan[TRow]
	statement
}

func newQuery[TRow any]() *query[TRow] {
	return &query[TRow]{
		statement: statement{
			Params: make(map[string]any),
		},
	}
}

func (ctx *query[TRow]) ScanRows(rows *sql.Rows) (data []TRow, err error) {
	panic("not implemented")
}
