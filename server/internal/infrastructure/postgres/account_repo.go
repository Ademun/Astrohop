package postgres

import (
	"astrohop/pkg/db"
	"context"
)

type AccountRepo struct {
	m *db.Manager
}

func NewAccountRepo(m *db.Manager) *AccountRepo {
	return &AccountRepo{m: m}
}

func (r *AccountRepo) CreateAccount(ctx context.Context, key string) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `insert into application.accounts (secret_key) values (crypt($1, gen_salt('sha256crypt')))`, key)
	return err
}

func (r *AccountRepo) GetAccountIDByKey(ctx context.Context, key string) (int64, error) {
	var id int64
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `select account_id from application.accounts where secret_key = crypt($1, secret_key)`, key).Scan(&id)
	return id, err
}
