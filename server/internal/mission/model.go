package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"time"

	"github.com/google/uuid"
)

type Mission struct {
	MissionID uuid.UUID
	Data      *Data
	MapData   *MapData
	CreatedAt time.Time
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
	OID  int64
	Name string
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

type TaskResult struct {
	Progress taskProgress
	Payload  *MapData
	Error    error
}

type Task struct {
	MissionID uuid.UUID
	Data      *Data
}
