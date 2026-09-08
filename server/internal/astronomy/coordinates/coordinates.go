package coordinates

import (
	"math"
)

const RadToDeg = 180 / math.Pi
const DegToRad = math.Pi / 180
const Obliquity = 23.44 * DegToRad

type Equatorial struct {
	RA  float64 // In degrees
	Dec float64 // In degrees
}

func (eq Equatorial) ToHorizontal(lst float64, lat float64) Horizontal {
	h := (lst * 15) - eq.RA
	latRad := lat * DegToRad
	hRad := h * DegToRad
	decRad := eq.Dec * DegToRad

	sinAlt := math.Sin(decRad)*math.Sin(latRad) + math.Cos(decRad)*math.Cos(latRad)*math.Cos(hRad)
	sinAlt = math.Max(-1.0, math.Min(1.0, sinAlt))
	altRad := math.Asin(sinAlt)
	y := -1 * math.Sin(hRad) * math.Cos(decRad)
	x := math.Sin(decRad) - math.Sin(latRad)*sinAlt

	azRad := math.Atan2(y, x)

	return Horizontal{
		Alt: altRad * RadToDeg,
		Az:  NormDeg(azRad * RadToDeg),
	}
}

type Horizontal struct {
	Alt float64 `json:"alt"` // In degrees
	Az  float64 `json:"az"`  // In degrees
}

func (hr Horizontal) ToEquatorial(lst float64, lat float64) Equatorial {
	latRad := lat * DegToRad
	altRad := hr.Alt * DegToRad
	azRad := hr.Az * DegToRad

	sinDec := math.Sin(altRad)*math.Sin(latRad) + math.Cos(altRad)*math.Cos(latRad)*math.Cos(azRad)
	sinDec = math.Max(-1.0, math.Min(1.0, sinDec))
	decRad := math.Asin(sinDec)
	y := -1 * math.Sin(azRad) * math.Cos(altRad)
	x := math.Sin(altRad) - sinDec*math.Sin(latRad)

	hRad := math.Atan2(y, x)
	h := hRad * RadToDeg / 15
	ra := NormDeg((lst - h) * 15)

	return Equatorial{
		RA:  ra,
		Dec: decRad * RadToDeg,
	}
}

type Ecliptic struct {
	Lat  float64 // In degrees
	Long float64 // In degrees
}

func (ec Ecliptic) ToEquatorial() Equatorial {
	latRad := ec.Lat * DegToRad
	longRad := ec.Long * DegToRad

	sinDec := math.Sin(latRad)*math.Cos(Obliquity) + math.Cos(latRad)*math.Sin(Obliquity)*math.Sin(longRad)
	sinDec = math.Max(-1.0, math.Min(1.0, sinDec))
	decRad := math.Asin(sinDec)

	y := math.Sin(longRad)*math.Cos(Obliquity) - math.Tan(latRad)*math.Sin(Obliquity)
	x := math.Cos(longRad)

	raRad := math.Atan2(y, x)

	return Equatorial{
		RA:  NormDeg(raRad * RadToDeg),
		Dec: decRad * RadToDeg,
	}
}

type GeoLocation struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func DistanceEq(a, b Equatorial) float64 {
	raaRad := a.RA * DegToRad
	decaRad := a.Dec * DegToRad
	rabRad := b.RA * DegToRad
	decbRad := b.Dec * DegToRad

	cosDist := math.Sin(decaRad)*math.Sin(decbRad) + math.Cos(decaRad)*math.Cos(decbRad)*math.Cos(raaRad-rabRad)
	cosDist = math.Max(-1.0, math.Min(1.0, cosDist))

	return math.Acos(cosDist) * RadToDeg
}

func NormDeg(deg float64) float64 {
	d := math.Mod(deg, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}
