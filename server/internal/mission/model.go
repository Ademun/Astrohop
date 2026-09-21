package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"time"

	"github.com/google/uuid"
)

type Mission struct {
	MissionID         uuid.UUID
	AccountID         int64
	Information       Information
	Data              Data
	DataVersion       int
	Map               *Map
	SourceDataVersion *int
	Overrides         *Overrides
	IsPublic          bool
	CreatedAt         time.Time
	ModifiedAt        *time.Time
}

type Information struct {
	Name        *string
	Description *string
}

type Data struct {
	Location   coordinates.GeoLocation
	Time       time.Time
	Objectives []Objective
	Conditions Conditions
}

type Map struct {
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

type Overrides struct {
}

type action int

const (
	actionView action = iota
	actionEdit
)

type taskProgress string

const (
	taskBuildingRoute taskProgress = "building_route"
	taskDone          taskProgress = "done"
	taskFailed        taskProgress = "failed"
)

type TaskResult struct {
	Progress taskProgress
	Payload  *Map
	Error    error
}

type Task struct {
	MissionID   uuid.UUID
	DataVersion int
	Data        *Data
}
