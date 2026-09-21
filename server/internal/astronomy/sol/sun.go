package sol

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"math"
	"time"
)

func CalculateSunPosition(t time.Time) (*coordinates.Ecliptic, float64) {
	jd := atime.UTCToJulian(t)
	T := (jd - 2451545.0) / 36525.0
	T2 := T * T

	Lo := 280.46646 + 36000.76983*T + 0.0003032*T2
	M := 357.52911 + 35999.05029*T - 0.0001537*T2
	C := (1.914602-0.004817*T-0.000014*T2)*math.Sin(M*coordinates.DegToRad) + (0.019993-0.000101*T)*math.Sin(2*M*coordinates.DegToRad) + 0.000289*math.Sin(3*M*coordinates.DegToRad)
	trueLo := Lo + C
	trueM := M + C
	e := 0.016708634 - 0.000042037*T - 0.0000001267*T2
	R := (1.000001018 * (1 - e*e)) / (1 + e*math.Cos(trueM*coordinates.DegToRad))
	omega := 125.04 - 1934.136*T
	apparentLo := trueLo - 0.00569 - 0.00478*math.Sin(omega*coordinates.DegToRad)
	return &coordinates.Ecliptic{
		Long: coordinates.NormDeg(apparentLo),
		Lat:  0, // Actually it is not exactly zero due to the action of moon and planets. However, such precision is not relevant in this application
	}, R
}
