package coordinates

import (
	"math"
)

const RadToDeg = 180 / math.Pi
const DegToRad = math.Pi / 180

type Equatorial struct {
	RA  float64 // In degrees
	Dec float64 // In degrees
}

func (eq Equatorial) ToHorizontal(lst float64, lat float64) Horizontal {
	h := (lst * 15) - eq.RA
	latRad := lat * DegToRad
	hRad := h * DegToRad
	decRad := eq.Dec * DegToRad

	altRad := math.Asin(math.Sin(decRad)*math.Sin(latRad) + math.Cos(decRad)*math.Cos(latRad)*math.Cos(hRad))
	y := -1 * math.Sin(hRad) * math.Cos(decRad) / math.Cos(altRad)
	x := (math.Sin(decRad) - math.Sin(latRad)*math.Sin(altRad)) / (math.Cos(latRad) * math.Cos(altRad))
	azRad := math.Atan2(y, x)
	if azRad < 0 {
		azRad += math.Pi * 2
	}

	return Horizontal{
		Alt: altRad * RadToDeg,
		Az:  azRad * RadToDeg,
	}
}

type Horizontal struct {
	Alt float64
	Az  float64
}

func (hr Horizontal) ToEquatorial(lst float64, lat float64) Equatorial {
	latRad := lat * DegToRad
	altRad := hr.Alt * DegToRad
	azRad := hr.Az * DegToRad

	decRad := math.Asin(math.Sin(altRad)*math.Sin(latRad) + math.Cos(altRad)*math.Cos(latRad)*math.Cos(azRad))
	y := -1 * math.Sin(azRad) * math.Cos(altRad) / math.Cos(decRad)
	x := (math.Sin(altRad) - math.Sin(decRad)*math.Sin(latRad)) / (math.Cos(decRad) * math.Cos(latRad))
	hRad := math.Atan2(y, x)
	if hRad < 0 {
		hRad += math.Pi * 2
	}
	h := hRad * RadToDeg / 15
	ra := lst - h
	return Equatorial{
		RA:  ra * 15,
		Dec: decRad * RadToDeg,
	}
}

type GeoLocation struct {
	Lat  float64 `json:"lat" binding:"required"`
	Long float64 `json:"long" binding:"required"`
}

func DistanceEq(a, b Equatorial) float64 {
	raaRad := a.RA * DegToRad
	decaRad := a.Dec * DegToRad
	rabRad := b.RA * DegToRad
	decbRad := b.Dec * DegToRad
	return math.Acos(math.Sin(decaRad)*math.Sin(decbRad)+math.Cos(decaRad)*math.Cos(decbRad)*math.Cos(raaRad-rabRad)) * RadToDeg
}
