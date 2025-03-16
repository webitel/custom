package postgres

import (
	"context"
	"math/rand"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type cluster struct {
	dc []*pgxpool.Pool
	dt *pgtype.Map
}

func (c *cluster) types() *pgtype.Map {
	if c.dt == nil {
		_ = c.primary().AcquireFunc(
			context.TODO(), func(conn *pgxpool.Conn) error {
				c.dt = conn.Conn().TypeMap()
				return nil
			},
		)
		if c.dt == nil {
			c.dt = pgtype.NewMap()
		}
	}
	return c.dt
}

// random defines how to create the random number.
func random(min, max int) int {
	// // rand.Seed: ensures that the number that is generated is random(almost).
	// rand.Seed(model.CurrentTime().UnixNano())
	return rand.Intn((max - min)) + min
}

func (c *cluster) primary() *pgxpool.Pool {
	return c.dc[0]
}

func (c *cluster) secondary() *pgxpool.Pool {
	// first is [primary]
	if n := len(c.dc); n > 1 {
		return c.dc[random(1, n)]
	}
	return c.primary()
}
