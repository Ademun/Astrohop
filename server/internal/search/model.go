package search

import "astrohop/internal/astronomy/coordinates"

type Object struct {
	OID        int64  `db:"oid"`
	Identifier string `db:"identifier"`
	ObjectType string `db:"object_type"`
}

type ObjectStellarData struct {
	EqCoords          coordinates.Equatorial
	ApparentMagnitude float32
}

type NavData struct {
	RA                float64 `db:"ra"`
	Dec               float64 `db:"dec"`
	ApparentMagnitude float32 `db:"apparent_mag"`
}
