package postgres

import (
	"astrohop/internal/account"
	"astrohop/pkg/db"
	"context"
	"uuid"
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

func (r *AccountRepo) GetMissionPermissions(ctx context.Context, accountID int64, missionID uuid.UUID) (*account.MissionPermissions, error) {
	var permissions account.MissionPermissions
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `
select
(a.account_id = $1) as owner,
m.is_public
from application.missions m
inner join application.accounts a on a.account_id = m.account_id
where mission_id = $2
`, accountID, missionID).Scan(&permissions.IsOwner, &permissions.IsMissionPublic)
	return &permissions, err
}
