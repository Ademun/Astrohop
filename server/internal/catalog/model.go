package catalog

import "astrohop/internal/astronomy/coordinates"

type ObjectID int64

type Class string

const (
	ClassStar  Class = "star"
	ClassDSO   Class = "dso"
	ClassOther Class = "other"
)

// ObjectPosition is a minimal projection used for routing and rendering.
type ObjectPosition struct {
	ID       ObjectID
	Position coordinates.Equatorial
}

// ObjectRef is a lightweight reference to an object for lists and reference data.
type ObjectRef struct {
	ID         ObjectID
	Class      Class
	CommonName *string
}

// SearchHit represents a search result found by collection identifiers.
// The same object may be found by several identifiers and appear in the results several times.
type SearchHit struct {
	ObjectRef
	Identifier string // Matched identifier in its original form, e.g. "M31".
	Collection string
}

// Metadata contains data intended for display to the user.
type Metadata struct {
	TypeName    string
	CommonName  *string
	Description *string
	ImageURL    *string
}

type StarData struct {
	VisualMag     float64
	SpectralClass *string
}

type DSOData struct {
	VisualMag *float64
	MajorAxis *float64
	MinorAxis *float64
	PosAngle  *float64
}

// Object contains complete information about an object.
// Depending on Class, at most one of Star or DSO is populated.
type Object struct {
	ID       ObjectID
	Class    Class
	Position coordinates.Equatorial
	Distance *float64 // Distance in parsecs; nil if the distance is unknown.
	Metadata Metadata
	Star     *StarData
	DSO      *DSOData
}

type CollectionID int64

type Collection struct {
	ID          CollectionID
	Name        string
	Description *string
}

// CollectionMember is an object as it appears within a collection.
type CollectionMember struct {
	ObjectRef
	Identifier string // Identifier within the collection, e.g. "M31".
}

// BoundarySegment represents a segment of a constellation boundary (IAU).
type BoundarySegment struct {
	Start coordinates.Equatorial
	End   coordinates.Equatorial
}

type Constellation struct {
	ID       int16
	IAU      string
	Common   string
	Boundary []BoundarySegment
	// Pattern is the stick figure: a set of polylines, each a sequence of stars.
	Pattern [][]ObjectPosition
}

type PageRequest struct {
	Limit  int // 0 means the default page size.
	Offset int
}

type Page[T any] struct {
	Items []T
	Total int
}
