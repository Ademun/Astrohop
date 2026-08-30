package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var txKey = "tx"

type Manager struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{
		pool: pool,
	}
}

func (m *Manager) WithTx(ctx context.Context, f func(context.Context) error) error {
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return f(ctx)
	}
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx manager: begin: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(err, pgx.ErrTxClosed) {
				err = errors.Join(err, fmt.Errorf("tx manager: rollback: %w", rbErr))
			}
			return
		}
		if cErr := tx.Commit(ctx); cErr != nil {
			err = fmt.Errorf("tx manager: commit: %w", cErr)
		}
	}()
	err = f(context.WithValue(ctx, txKey, tx))
	return err
}

func (m *Manager) GetExecutor(ctx context.Context) Executor {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return m.pool
	}
	return tx
}

/*func WithTxResult[T any](ctx context.Context, m *Manager, f func(context.Context) (T, error)) (T, error) {
	var result T
	err := m.WithTx(ctx, func(c context.Context) error {
		var inErr error
		result, inErr = f(c)
		return inErr
	})
	return result, err
}*/
