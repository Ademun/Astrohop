package coordinates

import "math"

func ParallacticAngle(bodyPos *Equatorial, lst float64, lat float64) float64 {
	h := NormDeg(lst*15 - bodyPos.RA)
	hRad, latRad, decRad := h*DegToRad, lat*DegToRad, bodyPos.Dec*DegToRad
	if cmp(h, 0, 0.001) && cmp(bodyPos.Dec, lat, 0.001) {
		return 0
	}
	y := math.Sin(hRad)
	x := math.Tan(latRad)*math.Cos(decRad) - math.Sin(decRad)*math.Cos(hRad)
	return math.Atan2(y, x) * RadToDeg
}

func NormDeg(deg float64) float64 {
	d := math.Mod(deg, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}

func NormHour(hour float64) float64 {
	d := math.Mod(hour, 24.0)
	if d < 0 {
		d += 24.0
	}
	return d
}

func cmp(a, b, k float64) bool {
	return math.Abs(a-b) < k
}
