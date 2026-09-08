package sol

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"math"
	"time"
)

func CalculateMoonPosition(t time.Time) (*coordinates.Ecliptic, float64) {
	jd := atime.UTCToJulian(t)
	T := (jd - 2451545.0) / 36525.0
	T2 := T * T
	T3 := T2 * T
	T4 := T3 * T

	Lp := 218.3164477 + 481267.88123421*T - 0.0015786*T2 + T3/538841 - T4/65194000
	D := 297.8501921 + 445267.1114034*T - 0.0018819*T2 + T3/545868 - T4/113065000
	M := 357.5291092 + 35999.0502909*T - 0.0001536*T2 + T3/24490000
	Mp := 134.9633964 + 477198.8675055*T + 0.0087414*T2 + T3/69699 - T4/14712000
	F := 93.2720950 + 483202.0175233*T - 0.0036539*T2 - T3/3526000 + T4/863310000

	Lp = coordinates.NormDeg(Lp)
	D = coordinates.NormDeg(D)
	M = coordinates.NormDeg(M)
	Mp = coordinates.NormDeg(Mp)
	F = coordinates.NormDeg(F)

	var lSum, rSum, bSum float64
	for _, row := range moonLngDistTable {
		d, m, mp, f, lc, rc := row[0], row[1], row[2], row[3], row[4], row[5]
		theta := (d*D + m*M + mp*Mp + f*F) * coordinates.DegToRad
		lSum += lc * math.Sin(theta)
		rSum += rc * math.Cos(theta)
	}

	for _, row := range moonLatTable {
		d, m, mp, f, bc := row[0], row[1], row[2], row[3], row[4]
		theta := (d*D + m*M + mp*Mp + f*F) * coordinates.DegToRad
		bSum += bc * math.Sin(theta)
	}

	long := Lp + lSum/1e6
	lat := bSum / 1e6
	dist := 385000.56 + rSum/1000.0

	return &coordinates.Ecliptic{
		Lat:  lat,
		Long: coordinates.NormDeg(long),
	}, dist
}

// D	M	M'	F	l (10⁻⁶°)	r (10⁻³ km)
var moonLngDistTable = [][]float64{
	{0, 0, 1, 0, 6288774, -20905355},
	{2, 0, -1, 0, 1274027, -3699111},
	{2, 0, 0, 0, 658314, -2955968},
	{0, 0, 2, 0, 213618, -569925},
	{0, 1, 0, 0, -185116, 48888},
	{0, 0, 0, 2, -114332, -3149},
	{2, 0, -2, 0, 58793, 246158},
	{2, 0, -1, 1, 56872, -152138},
	{2, 0, 0, 1, 53240, -170733},
	{0, 0, 2, 1, 34690, -204586},
}

// D	M	M'	F	b (10⁻⁶°)
var moonLatTable = [][]float64{
	{0, 0, 0, 1, 5128122},
	{0, 0, 1, 1, 280602},
	{0, 0, 1, -1, 277693},
	{2, 0, 0, -1, 173237},
	{2, 0, -1, 1, 55413},
	{2, 0, -1, -1, 46271},
	{2, 0, 0, 1, 32573},
	{0, 0, 2, 1, 17198},
}
