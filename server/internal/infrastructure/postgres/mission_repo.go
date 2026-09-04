package postgres

import (
	"astrohop/internal/mission"
	"astrohop/pkg/db"
	"context"

	"github.com/google/uuid"
)

type MissionRepo struct {
	m *db.Manager
}

func NewMissionRepo(m *db.Manager) *MissionRepo {
	return &MissionRepo{m: m}
}

func (r *MissionRepo) CreateMission(ctx context.Context, data *mission.Data, accountID int64) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `insert into application.missions (account_id, data) values ($1, $2) returning mission_id`, accountID, data).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *MissionRepo) GetMission(ctx context.Context, id uuid.UUID) (*mission.Mission, error) {
	var m mission.Mission
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `select mission_id, data, map_data, created_at from application.missions where mission_id=$1`, id).Scan(&m.MissionID, &m.Data, &m.MapData, &m.CreatedAt)
	return &m, err
}

func (r *MissionRepo) GetAccountMissions(ctx context.Context, accountID int64) ([]mission.Mission, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `select mission_id, created_at from application.missions where account_id=$1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	missions := make([]mission.Mission, 0)
	for rows.Next() {
		var m mission.Mission
		err := rows.Scan(&m.MissionID, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		missions = append(missions, m)
	}
	return missions, rows.Err()
}

func (r *MissionRepo) UpdateMissionData(ctx context.Context, missionID uuid.UUID, data *mission.Data) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `update application.missions set data = $1 where mission_id = $2`, data, missionID)
	return err
}

func (r *MissionRepo) UpdateMissionMapData(ctx context.Context, missionID uuid.UUID, data *mission.MapData) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `update application.missions set map_data = $1 where mission_id = $2`, data, missionID)
	return err
}

func (r *MissionRepo) UpdateMissionVisibility(ctx context.Context, missionID uuid.UUID, isPublic bool) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `update application.missions set is_public = $1 where mission_id = $2`, isPublic, missionID)
	return err
}

func (r *MissionRepo) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `delete from application.missions where mission_id = $1`, missionID)
	return err
}

func (r *MissionRepo) ValidateOwnership(ctx context.Context, missionID uuid.UUID, accountID int64) (bool, error) {
	var result bool
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `select (account_id = $1) from application.missions where mission_id = $2`, accountID, missionID).Scan(&result)
	return result, err
}

func (r *MissionRepo) IsPublic(ctx context.Context, missionID uuid.UUID) (bool, error) {
	var result bool
	err := r.m.GetExecutor(ctx).QueryRow(ctx, `select is_public from application.missions where mission_id = $1`, missionID).Scan(&result)
	return result, err
}
