package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/pkg/algo"
	"time"
	"uuid"
)

type Mission struct {
	MissionID uuid.UUID `db:"mission_id" json:"mission_id" binding:"required"`
	Data      *Data     `db:"data" json:"data" binding:"required"`
	MapData   *MapData  `db:"map_data" json:"map_data" binding:"required"`
	CreatedAt time.Time `db:"created_at" json:"created_at" binding:"required"`
}

type Data struct {
	Location   coordinates.GeoLocation `json:"location" binding:"required"`
	Time       time.Time               `json:"time" binding:"required"`
	Objectives []Objective             `json:"objectives" binding:"required"`
	Conditions Conditions              `json:"conditions" binding:"required"`
}

type MapData struct {
	Tour *algo.Tour `db:"tour" json:"tour" binding:"required"`
}

type Objective struct {
	OID  int64  `db:"oid" json:"oid" binding:"required"`
	Name string `db:"name" json:"name" binding:"required"`
}

type Conditions struct {
	LimitingMagnitude float32 `json:"limiting_magnitude" binding:"required"`
}

type taskProgress string

const (
	taskBuildingRoute taskProgress = "building_route"
	taskDone          taskProgress = "done"
	taskFailed        taskProgress = "failed"
)

type taskResult struct {
	Progress taskProgress
	Payload  *MapData
	Error    error
}

type missionTask struct {
	MissionID uuid.UUID
	Data      *Data
}
