package planner

import (
	"astrohop/internal/astronomy/coordinates"
	"time"
)

type Mission struct {
	MissionID int64       `db:"mission_id" json:"mission_id"`
	Info      MissionInfo `db:"info" json:"info"`
}

type MissionInfo struct {
	Location   coordinates.GeoLocation `json:"location" binding:"required"`
	Time       time.Time               `json:"time" binding:"required"`
	Objectives []MissionObjective      `json:"objectives" binding:"required"`
	Conditions MissionConditions       `json:"conditions" binding:"required"`
}

type MissionObjective struct {
	OID  int64  `db:"oid" json:"oid" binding:"required"`
	Name string `db:"name" json:"name" binding:"required"`
}

type MissionConditions struct {
	LimitingMagnitude float32 `json:"limiting_magnitude" binding:"required"`
}

type MissionValidationInfo struct {
	MissionID  int64                       `db:"mission_id" json:"mission_id" binding:"required"`
	Objectives []MissionValidatedObjective `json:"objectives" binding:"required"`
}

type MissionValidatedObjective struct {
	Name              string          `json:"name" binding:"required"`
	ApparentMagnitude float32         `json:"apparent_magnitude" binding:"required"`
	Status            ObjectiveStatus `json:"status" binding:"required"`
}

type ObjectiveStatus string

const (
	ObjectiveStatusVisible   ObjectiveStatus = "Visible"
	ObjectiveStatusDim       ObjectiveStatus = "Dim"
	ObjectiveStatusInvisible ObjectiveStatus = "Invisible"
)

type MissionStatus string

const (
	MissionStatusDraft  MissionStatus = "draft"
	MissionStatusDone   MissionStatus = "done"
	MissionStatusFailed MissionStatus = "failed"
)

type taskProgress string

const (
	taskStarted       taskProgress = "started"
	taskFetchingData  taskProgress = "fetching_data"
	taskBuildingRoute taskProgress = "building_route"
	taskGeneratingMap taskProgress = "generating_map"
	taskDone          taskProgress = "done"
	taskFailed        taskProgress = "failed"
)

type taskResult struct {
	Progress taskProgress
	Payload  []byte
	Error    error
}

type missionTask struct {
	MissionID int64
	Result    chan taskResult
}
