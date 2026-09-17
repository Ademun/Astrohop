package postgres

import (
	"astrohop/internal/mission"
	"astrohop/pkg/db"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MissionRepo struct {
	m *db.Manager
}

func NewMissionRepo(m *db.Manager) *MissionRepo {
	return &MissionRepo{m: m}
}

type missionRow struct {
	MissionID uuid.UUID        `db:"mission_id"`
	Data      *mission.Data    `db:"data"`
	MapData   *mission.MapData `db:"map_data"`
	CreatedAt time.Time        `db:"created_at"`
}

func (r missionRow) ToDomain() *mission.Mission {
	return &mission.Mission{
		MissionID: r.MissionID,
		Data:      r.Data,
		MapData:   r.MapData,
		CreatedAt: r.CreatedAt,
	}
}

func (r missionRow) FromDomain(m *mission.Mission) {
	r.MissionID = m.MissionID
	r.Data = m.Data
	r.MapData = m.MapData
	r.CreatedAt = m.CreatedAt
}

type missionListRow struct {
	MissionID uuid.UUID `db:"mission_id"`
	CreatedAt time.Time `db:"created_at"`
}

func (r missionListRow) ToDomain() mission.Mission {
	return mission.Mission{
		MissionID: r.MissionID,
		CreatedAt: r.CreatedAt,
	}
}

type missionIDRow struct {
	MissionID uuid.UUID `db:"mission_id"`
}

type missionOwnershipRow struct {
	IsOwner bool `db:"is_owner"`
}

type missionPublicRow struct {
	IsPublic bool `db:"is_public"`
}

func (r *MissionRepo) CreateMission(ctx context.Context, data *mission.Data, accountID int64) (uuid.UUID, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		insert into application.missions (account_id, data)
		values ($1, $2)
		returning mission_id
	`, accountID, data)
	if err != nil {
		return uuid.Nil, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[missionIDRow])
	if err != nil {
		return uuid.Nil, err
	}

	return row.MissionID, nil
}

func (r *MissionRepo) GetMission(ctx context.Context, id uuid.UUID) (*mission.Mission, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select mission_id, data, map_data, created_at
		from application.missions
		where mission_id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[missionRow])
	if err != nil {
		return nil, err
	}

	return row.ToDomain(), nil
}

func (r *MissionRepo) GetAccountMissions(ctx context.Context, accountID int64) ([]mission.Mission, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select mission_id, created_at
		from application.missions
		where account_id = $1
	`, accountID)
	if err != nil {
		return nil, err
	}

	rowMissions, err := pgx.CollectRows(rows, pgx.RowToStructByName[missionListRow])
	if err != nil {
		return nil, err
	}

	missions := make([]mission.Mission, len(rowMissions))
	for i := range rowMissions {
		missions[i] = rowMissions[i].ToDomain()
	}

	return missions, nil
}

func (r *MissionRepo) UpdateMissionData(ctx context.Context, missionID uuid.UUID, data *mission.Data) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `
		update application.missions
		set data = $1
		where mission_id = $2
	`, data, missionID)
	return err
}

func (r *MissionRepo) UpdateMissionMapData(ctx context.Context, missionID uuid.UUID, data *mission.MapData) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `
		update application.missions
		set map_data = $1
		where mission_id = $2
	`, data, missionID)
	return err
}

func (r *MissionRepo) UpdateMissionVisibility(ctx context.Context, missionID uuid.UUID, isPublic bool) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `
		update application.missions
		set is_public = $1
		where mission_id = $2
	`, isPublic, missionID)
	return err
}

func (r *MissionRepo) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.m.GetExecutor(ctx).Exec(ctx, `
		delete from application.missions
		where mission_id = $1
	`, missionID)
	return err
}

func (r *MissionRepo) ValidateOwnership(ctx context.Context, missionID uuid.UUID, accountID int64) (bool, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select (account_id = $1) as is_owner
		from application.missions
		where mission_id = $2
	`, accountID, missionID)
	if err != nil {
		return false, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[missionOwnershipRow])
	if err != nil {
		return false, err
	}

	return row.IsOwner, nil
}

func (r *MissionRepo) IsPublic(ctx context.Context, missionID uuid.UUID) (bool, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
		select is_public
		from application.missions
		where mission_id = $1
	`, missionID)
	if err != nil {
		return false, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[missionPublicRow])
	if err != nil {
		return false, err
	}

	return row.IsPublic, nil
}
