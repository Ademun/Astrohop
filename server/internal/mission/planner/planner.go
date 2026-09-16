package planner

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/astronomy/sol"
	"astrohop/internal/search"
	"astrohop/pkg/algo"
)

func Build(in *Input) *Output {
	lst := atime.GetLocalSidereal(in.Location, in.Time)

	oids := make([]int64, len(in.Objectives))
	stellar := make([]search.ObjectStellarData, len(in.Objectives))
	positions := make(map[int64]coordinates.Horizontal, len(in.Objectives))
	for i, o := range in.Objectives {
		oids[i] = o.OID
		stellar[i] = o.Stellar
		positions[o.OID] = o.Stellar.EqCoords.ToHorizontal(lst, in.Location.Lat)
	}

	tour := buildTour(oids, stellar)

	moonPosition, _ := sol.CalculateMoonPosition(in.Time)

	return &Output{
		MoonPosition: moonPosition.ToEquatorial().ToHorizontal(lst, in.Location.Lat),
		Positions:    positions,
		Tour:         tour,
	}
}

func buildTour(objectOids []int64, stellarData []search.ObjectStellarData) []int64 {
	distanceMtrx := make([][]float64, len(stellarData))
	for i := range stellarData {
		distanceMtrx[i] = make([]float64, len(stellarData))
	}
	for i := 0; i < len(stellarData)-1; i++ {
		for j := i + 1; j < len(stellarData); j++ {
			dist := coordinates.DistanceEq(stellarData[i].EqCoords, stellarData[j].EqCoords)
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
