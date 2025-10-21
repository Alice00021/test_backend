package transactional

import (
	"context"
	"fmt"
	"test_go/pkg/postgres"
)

type Transactional interface {
	RunInTransaction(context.Context, func(tCtx context.Context) error) error
}

type pgTransaction struct {
	pg *postgres.Postgres
}

func NewPgTransaction(pg *postgres.Postgres) Transactional {
	return &pgTransaction{pg: pg}
}

func (t *pgTransaction) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := t.pg.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	var commited = false
	defer func() {
		if !commited {
			err := tx.Rollback(ctx)
			if err != nil {
				_ = fmt.Errorf("error rollback %w", err)
			}
		}
	}()

	txCtx := context.WithValue(ctx, postgres.TXClientContextKey, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	commited = true
	return nil
}
