package catalog

import (
	"astrohop/internal/astronomy/coordinates"
)

type ObjectRefDTO struct {
	ID         int64   `json:"id"`
	Class      string  `json:"class"`
	CommonName *string `json:"common_name"`
}

type CollectionDTO struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type CollectionMemberDTO struct {
	ObjectRefDTO
	Identifier string `json:"identifier"`
}

type SearchHitDTO struct {
	ObjectRefDTO
	Identifier string `json:"identifier"`
	Collection string `json:"collection"`
}

type EquatorialDTO struct {
	RA  float64 `json:"ra"`
	Dec float64 `json:"dec"`
}

type MetadataDTO struct {
	TypeName    string  `json:"type_name"`
	CommonName  *string `json:"common_name"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
}

type StarDataDTO struct {
	VisualMag     float64 `json:"visual_mag"`
	SpectralClass *string `json:"spectral_class"`
}

type DSODataDTO struct {
	VisualMag *float64 `json:"visual_mag"`
	MajorAxis *float64 `json:"major_axis"`
	MinorAxis *float64 `json:"minor_axis"`
	PosAngle  *float64 `json:"pos_angle"`
}

type ObjectDTO struct {
	ID       int64         `json:"id"`
	Class    string        `json:"class"`
	Position EquatorialDTO `json:"position"`
	Distance *float64      `json:"distance_pc"`
	Metadata MetadataDTO   `json:"metadata"`
	Star     *StarDataDTO  `json:"star,omitempty"`
	DSO      *DSODataDTO   `json:"dso,omitempty"`
}

type SearchRequestDTO struct {
	Query      *string `form:"query"`
	Collection *int64  `form:"collection"`
	Limit      int     `form:"limit"`
	Offset     int     `form:"offset"`
}

func (r SearchRequestDTO) ToDomain() SearchRequest {
	return SearchRequest{Query: r.Query, CollectionID: (*CollectionID)(r.Collection), Limit: r.Limit, Offset: r.Offset}
}

func newObjectRefDTO(r ObjectRef) ObjectRefDTO {
	return ObjectRefDTO{
		ID:         int64(r.ID),
		Class:      string(r.Class),
		CommonName: r.CommonName,
	}
}

func newCollectionDTO(c Collection) CollectionDTO {
	return CollectionDTO{
		ID:          int64(c.ID),
		Name:        c.Name,
		Description: c.Description,
	}
}

func newSearchHitDTO(h SearchHit) SearchHitDTO {
	return SearchHitDTO{
		ObjectRefDTO: newObjectRefDTO(h.ObjectRef),
		Identifier:   h.Identifier,
		Collection:   h.Collection,
	}
}

func newEquatorialDTO(e coordinates.Equatorial) EquatorialDTO {
	return EquatorialDTO{RA: e.RA, Dec: e.Dec}
}

func newObjectDTO(o Object) ObjectDTO {
	dto := ObjectDTO{
		ID:       int64(o.ID),
		Class:    string(o.Class),
		Position: newEquatorialDTO(o.Position),
		Distance: o.Distance,
		Metadata: MetadataDTO{
			TypeName:    o.Metadata.TypeName,
			CommonName:  o.Metadata.CommonName,
			Description: o.Metadata.Description,
			ImageURL:    o.Metadata.ImageURL,
		},
	}
	if o.Star != nil {
		dto.Star = &StarDataDTO{
			VisualMag:     o.Star.VisualMag,
			SpectralClass: o.Star.SpectralClass,
		}
	}
	if o.DSO != nil {
		dto.DSO = &DSODataDTO{
			VisualMag: o.DSO.VisualMag,
			MajorAxis: o.DSO.MajorAxis,
			MinorAxis: o.DSO.MinorAxis,
			PosAngle:  o.DSO.PosAngle,
		}
	}
	return dto
}
