package planner

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/catalog"
	"time"
)

type Objective struct {
	OID      int64
	Position catalog.ObjectPosition
}

type Input struct {
	Location   coordinates.GeoLocation
	Time       time.Time
	Objectives []Objective
}

type Output struct {
	MoonPosition coordinates.Horizontal
	Positions    map[int64]coordinates.Horizontal
	Tour         []int64
}
