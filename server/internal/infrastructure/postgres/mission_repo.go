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

func (r *MissionRepo) CreateMission(ctx context.Context, accountID int64, data *mission.Data) (uuid.UUID, error) {
	exec := r.m.GetExecutor(ctx)
	var missionID uuid.UUID
	err := exec.QueryRow(ctx,
		`insert into application.missions (account_id, data) values ($1, $2) returning mission_id`,
		accountID, data).Scan(&missionID)
	return missionID, err
}

func (r *MissionRepo) GetMission(ctx context.Context, missionID uuid.UUID) (*mission.Mission, error) {
	exec := r.m.GetExecutor(ctx)
	rows, err := exec.Query(ctx, `
		select m.mission_id,
		       m.account_id,
		       m.name,
		       m.description,
		       m.data,
		       m.data_version,
		       mm.map,
		       mm.source_data_version,
		       mm.overrides,
		       m.is_public,
		       m.created_at,
		       m.modified_at
		from application.missions m
		         left join application.mission_maps mm on m.mission_id = mm.mission_id
		where m.mission_id = $1`, missionID)
	if err != nil {
		return nil, err
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[missionWithMapRow])
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *MissionRepo) GetAccountMissions(ctx context.Context, accountID int64) ([]mission.Mission, error) {
	exec := r.m.GetExecutor(ctx)
	rows, err := exec.Query(ctx, `
		select mission_id, name, description, is_public, created_at, modified_at
		from application.missions
		where account_id = $1
		order by created_at desc`, accountID)
	if err != nil {
		return nil, err
	}
	rowsData, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[missionSummaryRow])
	if err != nil {
		return nil, err
	}
	missions := make([]mission.Mission, len(rowsData))
	for i := range rowsData {
		missions[i] = rowsData[i].toDomain()
	}
	return missions, nil
}

func (r *MissionRepo) GetMissionOwner(ctx context.Context, missionID uuid.UUID) (int64, error) {
	exec := r.m.GetExecutor(ctx)
	var accountID int64
	err := exec.QueryRow(ctx,
		`select account_id from application.missions where mission_id = $1`,
		missionID).Scan(&accountID)
	return accountID, err
}

func (r *MissionRepo) UpdateMissionInformation(ctx context.Context, info *mission.Information, missionID uuid.UUID) error {
	exec := r.m.GetExecutor(ctx)
	_, err := exec.Exec(ctx,
		`update application.missions set name = $1, description = $2 where mission_id = $3`,
		info.Name, info.Description, missionID)
	return err
}

func (r *MissionRepo) UpdateMissionData(ctx context.Context, data *mission.Data, missionID uuid.UUID) (int, error) {
	exec := r.m.GetExecutor(ctx)
	var dataVersion int
	err := exec.QueryRow(ctx,
		`update application.missions
		 set data = $1
		 where mission_id = $2
		 returning data_version`,
		data, missionID).Scan(&dataVersion)
	return dataVersion, err
}

func (r *MissionRepo) UpdateMissionMap(ctx context.Context, sourceDataVersion int, m *mission.Map, missionID uuid.UUID) error {
	exec := r.m.GetExecutor(ctx)
	_, err := exec.Exec(ctx, `
		insert into application.mission_maps (mission_id, source_data_version, map)
		values ($1, $2, $3)
		on conflict (mission_id) do update set source_data_version = excluded.source_data_version,
		                                       map                 = excluded.map
		where application.mission_maps.source_data_version < excluded.source_data_version`,
		missionID, sourceDataVersion, m)
	return err
}

func (r *MissionRepo) SetMissionMapOverrides(ctx context.Context, overrides *mission.Overrides, missionID uuid.UUID) error {
	exec := r.m.GetExecutor(ctx)
	_, err := exec.Exec(ctx,
		`update application.mission_maps set overrides = $1 where mission_id = $2`,
		overrides, missionID)
	return err
}

func (r *MissionRepo) SetMissionVisibility(ctx context.Context, isPublic bool, missionID uuid.UUID) error {
	exec := r.m.GetExecutor(ctx)
	_, err := exec.Exec(ctx,
		`update application.missions set is_public = $1 where mission_id = $2`,
		isPublic, missionID)
	return err
}

func (r *MissionRepo) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	exec := r.m.GetExecutor(ctx)
	_, err := exec.Exec(ctx,
		`delete from application.missions where mission_id = $1`,
		missionID)
	return err
}

type missionRow struct {
	MissionID   uuid.UUID    `db:"mission_id"`
	AccountID   int64        `db:"account_id"`
	Name        *string      `db:"name"`
	Description *string      `db:"description"`
	Data        mission.Data `db:"data"`
	DataVersion int          `db:"data_version"`
	IsPublic    bool         `db:"is_public"`
	CreatedAt   time.Time    `db:"created_at"`
	ModifiedAt  *time.Time   `db:"modified_at"`
}

func (r missionRow) toDomain() mission.Mission {
	return mission.Mission{
		MissionID:   r.MissionID,
		AccountID:   r.AccountID,
		Information: mission.Information{Name: r.Name, Description: r.Description},
		Data:        r.Data,
		DataVersion: r.DataVersion,
		IsPublic:    r.IsPublic,
		CreatedAt:   r.CreatedAt,
		ModifiedAt:  r.ModifiedAt,
	}
}

type missionSummaryRow struct {
	MissionID   uuid.UUID  `db:"mission_id"`
	Name        *string    `db:"name"`
	Description *string    `db:"description"`
	IsPublic    bool       `db:"is_public"`
	CreatedAt   time.Time  `db:"created_at"`
	ModifiedAt  *time.Time `db:"modified_at"`
}

func (r missionSummaryRow) toDomain() mission.Mission {
	return mission.Mission{
		MissionID:   r.MissionID,
		Information: mission.Information{Name: r.Name, Description: r.Description},
		IsPublic:    r.IsPublic,
		CreatedAt:   r.CreatedAt,
		ModifiedAt:  r.ModifiedAt,
	}
}

type missionWithMapRow struct {
	missionRow
	Map               *mission.Map       `db:"map"`
	SourceDataVersion *int               `db:"source_data_version"`
	Overrides         *mission.Overrides `db:"overrides"`
}

func (r missionWithMapRow) toDomain() *mission.Mission {
	m := r.missionRow.toDomain()
	m.Map = r.Map
	m.SourceDataVersion = r.SourceDataVersion
	m.Overrides = r.Overrides
	return &m
}
