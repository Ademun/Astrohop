package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"time"

	"github.com/google/uuid"
)

type Mission struct {
	MissionID uuid.UUID `db:"mission_id"`
	Data      *Data     `db:"data"`
	MapData   *MapData  `db:"map_data"`
	CreatedAt time.Time `db:"created_at"`
}

type Data struct {
	Location   coordinates.GeoLocation
	Time       time.Time
	Objectives []Objective
	Conditions Conditions
}

type MapData struct {
	MoonPosition coordinates.Horizontal
	Positions    map[int64]coordinates.Horizontal
	Tour         []int64
}

type Objective struct {
	OID  int64  `db:"oid"`
	Name string `db:"name"`
}

type Conditions struct {
	LimitingMagnitude float32
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
