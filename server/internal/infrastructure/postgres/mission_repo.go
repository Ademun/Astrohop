package postgres

import (
	"astrohop/internal/planner"
	"astrohop/pkg/db"
	"context"

	"github.com/jackc/pgx/v5"
)

type MissionRepo struct {
	m *db.Manager
}

func NewMissionRepo(m *db.Manager) *MissionRepo {
	return &MissionRepo{m: m}
}

func (r *MissionRepo) CreateNewMission(ctx context.Context, mission *planner.MissionInfo, accessToken string) (int64, error) {
	var id int64
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `insert into application.missions (info, access_token) values ($1, crypt($2, gen_salt('sha256crypt'))) returning mission_id`, mission, accessToken).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *MissionRepo) GetMission(ctx context.Context, missionId int64) (*planner.Mission, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `select mission_id, info from application.missions where mission_id = $1`, missionId)
	if err != nil {
		return nil, err
	}

	mission, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[planner.Mission])
	if err != nil {
		return nil, err
	}

	return &mission, nil
}

func (r *MissionRepo) ValidateOwnership(ctx context.Context, missionId int64, accessToken string) (bool, error) {
	var result bool
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `select (crypt($1, access_token) = access_token) from application.missions where mission_id = $2`, accessToken, missionId).Scan(&result)
	if err != nil {
		return false, err
	}
	return result, nil
}
