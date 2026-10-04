package coordinates

import "math"

const RadToDeg = 180 / math.Pi
const DegToRad = math.Pi / 180

func Rad(deg float64) float64 {
	return deg * DegToRad
}

func Deg(rad float64) float64 {
	return rad * DegToRad
}

func ParallacticAngle(bodyPos *Equatorial, lst float64, lat float64) float64 {
	h := NormRad(Rad(lst*15) - bodyPos.RA)
	if cmp(h, 0, 0.001) && cmp(bodyPos.Dec, lat, 0.001) {
		return 0
	}
	y := math.Sin(h)
	x := math.Tan(lat)*math.Cos(bodyPos.Dec) - math.Sin(bodyPos.Dec)*math.Cos(h)
	return math.Atan2(y, x)
}

func NormDeg(rad float64) float64 {
	d := math.Mod(rad, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}

func NormRad(rad float64) float64 {
	r := math.Mod(rad, math.Pi*2)
	if r < 0 {
		r += math.Pi * 2
	}
	return r
}

func NormHour(hour float64) float64 {
	d := math.Mod(hour, 24.0)
	if d < 0 {
		d += 24.0
	}
	return d
}

func Sign(n float64) float64 {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func cmp(a, b, k float64) bool {
	return math.Abs(a-b) < k
}
