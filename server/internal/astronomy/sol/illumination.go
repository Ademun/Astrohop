package sol

import "math"

const ZeroMagFluxVFreq = 3.64e-20

func VBandMagnitudeToFlux(magnitude float64) float64 {
	return ZeroMagFluxVFreq * math.Pow(10, -0.4*magnitude)
}

func VBandFluxToMagnitude(flux float64) float64 {
	return -2.5 * math.Log10(flux/ZeroMagFluxVFreq)
}
