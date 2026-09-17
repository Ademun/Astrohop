package postgres

import (
	"astrohop/internal/account"
	"astrohop/pkg/db"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AccountRepo struct {
	m *db.Manager
}

func NewAccountRepo(m *db.Manager) *AccountRepo {
	return &AccountRepo{m: m}
}

type accountIDRow struct {
	AccountID int64 `db:"account_id"`
}

func (r accountIDRow) ToDomain() int64 {
	return r.AccountID
}

type accountMissionPermissions struct {
	IsOwner         bool `db:"owner"`
	IsMissionPublic bool `db:"is_public"`
}

func (r accountMissionPermissions) ToDomain() account.MissionPermissions {
	return account.MissionPermissions{
		IsOwner:         r.IsOwner,
		IsMissionPublic: r.IsMissionPublic,
	}
}

func (r accountMissionPermissions) FromDomain(p account.MissionPermissions) {
	r.IsOwner = p.IsOwner
	r.IsMissionPublic = p.IsMissionPublic
}

func (r *AccountRepo) CreateAccount(ctx context.Context, key string) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `
		insert into application.accounts (secret_key)
		values (crypt($1, gen_salt('sha256crypt')))
	`, key)
	return err
}

func (r *AccountRepo) GetAccountIDByKey(ctx context.Context, key string) (int64, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select account_id
		from application.accounts
		where secret_key = crypt($1, secret_key)
	`, key)
	if err != nil {
		return 0, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[accountIDRow])
	if err != nil {
		return 0, err
	}

	return row.ToDomain(), nil
}

func (r *AccountRepo) GetMissionPermissions(ctx context.Context, accountID int64, missionID uuid.UUID) (*account.MissionPermissions, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select
			(a.account_id = $1) as owner,
			m.is_public
		from application.missions m
		inner join application.accounts a on a.account_id = m.account_id
		where mission_id = $2
	`, accountID, missionID)
	if err != nil {
		return nil, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[accountMissionPermissions])
	if err != nil {
		return nil, err
	}

	permissions := row.ToDomain()
	return &permissions, nil
}
