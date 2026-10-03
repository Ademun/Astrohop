package planner

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/astronomy/sol"
	"astrohop/internal/catalog"
	"astrohop/pkg/algo"
)

func Build(in *Input) *Output {
	lst := atime.GetLocalSidereal(in.Location, in.Time)

	oids := make([]int64, len(in.Objectives))
	positions := make([]catalog.ObjectPosition, len(in.Objectives))
	horizontal := make(map[int64]coordinates.Horizontal, len(in.Objectives))
	for i, o := range in.Objectives {
		oids[i] = o.OID
		positions[i] = o.Position
		horizontal[o.OID] = o.Position.Position.ToHorizontal(lst, in.Location.Lat)
	}

	tour := buildTour(oids, positions)

	moonData := sol.CalculateMoonPosition(in.Time)

	return &Output{
		MoonPosition: moonData.ToEquatorial().ToHorizontal(lst, in.Location.Lat),
		Positions:    horizontal,
		Tour:         tour,
	}
}

func buildTour(objectOids []int64, positions []catalog.ObjectPosition) []int64 {
	distanceMtrx := make([][]float64, len(positions))
	for i := range positions {
		distanceMtrx[i] = make([]float64, len(positions))
	}
	for i := 0; i < len(positions)-1; i++ {
		for j := i + 1; j < len(positions); j++ {
			dist := coordinates.DistanceEq(positions[i].Position, positions[j].Position)
			distanceMtrx[i][j] = dist
			distanceMtrx[j][i] = dist
		}
	}
	order := algo.UseFarthestInsertion(distanceMtrx)
	tour := make([]int64, len(objectOids))
	for i, o := range order {
		tour[i] = objectOids[o]
	}
	return tour
}
