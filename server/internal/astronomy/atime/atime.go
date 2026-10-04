package atime

import (
	"astrohop/internal/astronomy/coordinates"
	"math"
	"time"
)

// UTCToJulian converts a UTC time to a Julian Date (JD).
// Formula source:
// U.S. Naval Observatory, Astronomical Applications Department
// "Converting Between Julian Dates and Gregorian Calendar Dates"
// https://aa.usno.navy.mil/faq/JD_Formula
//
// The implemented formula:
//
//	JD = 367K - <(7(K+<(M+9)/12>))/4> + <(275M)/9> + I + 1721013.5 + UT1/24 - 0.5sign(100K+M-190002.5) + 0.5
func UTCToJulian(t time.Time) float64 {
	t = t.UTC()
	Ki, Mi, Ii := t.Date()
	K, M, I := float64(Ki), float64(Mi), float64(Ii)
	UT := float64(t.Hour()) + float64(t.Minute())/60 + float64(t.Second())/3600 + float64(t.Nanosecond())/3.6e12
	JD := 367*K -
		math.Trunc((7*(K+math.Trunc((M+9)/12)))/4) +
		math.Trunc((275*M)/9) +
		I +
		1721013.5 +
		UT/24 -
		0.5*coordinates.Sign(100*K+M-190002.5) + 0.5
	return JD
}

// UTCToGAST converts a UTC time to Greenwich Apparent Sidereal Time (GAST) in hours.
// Formula source:
// U.S. Naval Observatory, Astronomical Applications Department
// "Sidereal Time"
// https://aa.usno.navy.mil/faq/GAST
//
// The implemented algorithm:
//
//	GAST = GMST + Δψ·cos(ε)
//
// where GMST is Greenwich Mean Sidereal Time computed from the Julian Date,
// Δψ is the nutation in longitude, and ε is the obliquity of the ecliptic.
// The function uses UTC directly (approximating UT1 and TT).
func UTCToGAST(t time.Time) float64 {
	JDut := UTCToJulian(t)
	JD0 := math.Floor(JDut-0.5) + 0.5
	H := (JDut - JD0) * 24
	Dtt := JDut - 2451545.0
	Dut := JD0 - 2451545.0
	T := Dtt / 36525
	GMST := coordinates.NormHour(6.697375 + 0.065709824279*Dut + 1.0027379*H + 0.0000258*math.Pow(T, 2))
	omega := coordinates.Rad(125.04 - 0.052954*Dtt)
	L := coordinates.Rad(280.47 + 0.98565*Dtt)
	epsilon := coordinates.Rad(23.4393 - 0.0000004*Dtt)
	nutation := -0.000319*math.Sin(omega) - 0.000024*math.Sin(2*L)
	eqeq := nutation * math.Cos(epsilon)
	GAST := coordinates.NormHour(GMST + eqeq)
	return GAST
}

func GetLocalSidereal(loc coordinates.GeoLocation, time time.Time) float64 {
	gast := UTCToGAST(time)
	long := loc.Long / 15
	lst := coordinates.NormHour(gast + long)
	return lst
}
