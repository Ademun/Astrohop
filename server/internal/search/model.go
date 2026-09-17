package search

import "astrohop/internal/astronomy/coordinates"

type Object struct {
	OID        int64
	Identifier string
	ObjectType string
}

type ObjectStellarData struct {
	EqCoords          coordinates.Equatorial
	ApparentMagnitude float32
}

type NavData struct {
	RA                float64
	Dec               float64
	ApparentMagnitude float32
}
