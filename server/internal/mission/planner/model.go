package planner

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/search"
	"time"
)

type Objective struct {
	OID     int64
	Stellar search.ObjectStellarData
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
