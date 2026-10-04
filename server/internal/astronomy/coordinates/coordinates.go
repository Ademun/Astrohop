package coordinates

// All struct fields are expressed in radians and all calculations are performed in radians

import (
	"math"
)

const Obliquity = 23.44 * DegToRad

type Equatorial struct {
	RA  float64 // Right ascension in [0, 2π).
	Dec float64 // Declination in [-π/2, π/2].
}

func NewEquatorial(ra float64, dec float64) Equatorial {
	return Equatorial{ra, dec}
}

func (eq Equatorial) ToHorizontal(lst float64, lat float64) Horizontal {
	h := NormRad(Rad(lst*15) - eq.RA)

	sinAlt := math.Sin(eq.Dec)*math.Sin(lat) + math.Cos(eq.Dec)*math.Cos(lat)*math.Cos(h)
	sinAlt = math.Max(-1.0, math.Min(1.0, sinAlt))
	alt := math.Asin(sinAlt)
	y := -1 * math.Sin(h) * math.Cos(eq.Dec)
	x := math.Sin(eq.Dec) - math.Sin(lat)*sinAlt/math.Cos(lat)

	az := NormRad(math.Atan2(y, x))

	return Horizontal{
		Alt: alt,
		Az:  az,
	}
}

type Horizontal struct {
	Az  float64 `json:"az"`  // Azimuth in [0, 2π).
	Alt float64 `json:"alt"` // Altitude in [-π/2, π/2].
}

type Ecliptic struct {
	Long float64 // Celestial longitude in [0, 2π)
	Lat  float64 // Celestial latitude in [-π/2, π/2].
	Dist float64 // Km
}

func (ec Ecliptic) ToEquatorial() Equatorial {
	sinDec := math.Sin(ec.Lat)*math.Cos(Obliquity) + math.Cos(ec.Lat)*math.Sin(Obliquity)*math.Sin(ec.Long)
	sinDec = math.Max(-1.0, math.Min(1.0, sinDec))
	dec := math.Asin(sinDec)

	y := math.Sin(ec.Long)*math.Cos(Obliquity) - math.Tan(ec.Lat)*math.Sin(Obliquity)
	x := math.Cos(ec.Long)

	ra := NormRad(math.Atan2(y, x))

	return Equatorial{
		RA:  ra,
		Dec: dec,
	}
}

type GeoLocation struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func DistanceEq(a, b Equatorial) float64 {
	cosDist := math.Sin(a.Dec)*math.Sin(b.Dec) + math.Cos(a.Dec)*math.Cos(b.Dec)*math.Cos(a.RA-b.RA)
	cosDist = math.Max(-1.0, math.Min(1.0, cosDist))

	return math.Acos(cosDist)
}
