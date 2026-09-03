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
		0.5*sign(100*K+M-190002.5) + 0.5
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
	GMST := math.Mod(6.697375+0.065709824279*Dut+1.0027379*H+0.0000258*math.Pow(T, 2), 24)
	if GMST < 0 {
		GMST += 24
	}
	omega := (125.04 - 0.052954*Dtt) * (math.Pi / 180)
	L := (280.47 + 0.98565*Dtt) * (math.Pi / 180)
	epsilon := (23.4393 - 0.0000004*Dtt) * (math.Pi / 180)
	nutation := -0.000319*math.Sin(omega) - 0.000024*math.Sin(2*L)
	eqeq := nutation * math.Cos(epsilon)
	GAST := math.Mod(GMST+eqeq, 24)
	if GAST < 0 {
		GAST += 24
	}
	return GAST
}

func GetLocalSidereal(loc coordinates.GeoLocation, time time.Time) float64 {
	gast := UTCToGAST(time)
	long := loc.Long / 15
	lst := math.Mod(gast+long, 24)
	if lst < 0 {
		lst += 24
	}
	return lst
}

func sign(n float64) float64 {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
