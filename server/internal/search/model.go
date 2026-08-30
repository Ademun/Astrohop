package search

import "astrohop/internal/astronomy/coordinates"

type Object struct {
	OID        int64  `db:"oid" json:"oid" binding:"required"`
	Identifier string `db:"identifier" json:"name" binding:"required"`
	ObjectType string `db:"object_type" json:"type" binding:"required"`
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
