package repo

import (
	"context"
	"database/sql"
)

type noOpBeginTx struct {
	*sql.Tx
}

var _ Executor = noOpBeginTx{}

func (tx noOpBeginTx) BeginTx(
	context.Context,
	*sql.TxOptions,
) (*sql.Tx, error) {
	// intended no operation
	// it is here just to satisfy type system to allow '*sql.Tx' as an 'Executor' to pass to 'TxRepository'
	// but it is totally safe as the 'Repo' objest is downscaled to a 'Repository' object
	return nil, nil
}

func (repo *Repository) Transaction(
	ctx context.Context,
	fn func(context.Context, Service) error,
) (err error) {
	var tx *sql.Tx
	tx, err = repo.exec.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	r := NewRepository(noOpBeginTx{tx})
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = fn(ctx, r)
	return
}
