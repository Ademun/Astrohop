package postgres

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/catalog"
	"astrohop/pkg/db"
	"astrohop/pkg/utils"
	"context"
	"errors"
	"fmt"
)

type CatalogRepo struct {
	m *db.Manager
}

func NewCatalogRepo(m *db.Manager) *CatalogRepo {
	return &CatalogRepo{m: m}
}

func (r *CatalogRepo) Collections(ctx context.Context) ([]catalog.Collection, error) {
	rows, err := r.m.CollectRows[collectionRow](ctx, `
		select id, name, description
		from catalog.collections
		order by name
	`)
	if err != nil {
		return nil, err
	}
	return utils.Map(rows, collectionRow.toDomain), nil
}

func (r *CatalogRepo) Object(ctx context.Context, id catalog.ObjectID) (catalog.Object, error) {
	row, err := r.m.CollectOne[objectRow](ctx, `
		select * from catalog.object_full_v where id = $1
	`, id)
	if err != nil {
		return catalog.Object{}, err
	}
	return row.toDomain()
}

func (r *CatalogRepo) Positions(ctx context.Context, ids []catalog.ObjectID) ([]catalog.ObjectPosition, error) {
	rows, err := r.m.CollectRows[positionRow](ctx, `
		select p.id, p.ra_rad, p.dec_rad
		from unnest($1::bigint[]) with ordinality as req(id, ord)
		join catalog.navigation_position_v p on p.id = req.id
		order by req.ord
	`, ids)
	if err != nil {
		return nil, err
	}

	if len(rows) != len(ids) {
		found := make(map[int64]struct{}, len(rows))
		for _, row := range rows {
			found[row.ID] = struct{}{}
		}
		var missing []catalog.ObjectID
		for _, id := range ids {
			if _, ok := found[int64(id)]; !ok {
				missing = append(missing, id)
			}
		}
		return nil, fmt.Errorf("%w: %v", catalog.ErrNotFound, missing)
	}

	return utils.Map(rows, positionRow.toDomain), nil
}

func (r *CatalogRepo) Search(ctx context.Context, request catalog.SearchRequest) ([]catalog.SearchHit, error) {
	rows, err := r.m.CollectRows[searchRow](ctx, `
		with q as (
			select catalog.norm_ident($1) as v, catalog.norm_ident(catalog.escape_like($1)) as pat
		)
		select d.object_id as id, d.class, d.common_name, d.identifier, d.collection
		from catalog.collection_object_details_v d, q
		where ($1::text is null or d.identifier_norm like '%' || q.pat || '%')
		  and ($2::smallint is null or d.collection_id = $2::smallint)
		order by
			case
				when d.identifier_norm = q.v then 0
				when d.identifier_norm like q.pat || '%' then 1
				else 2
			end,
			length(d.identifier),
			d.identifier
		limit $3 offset $4
	`, request.Query, request.CollectionID, request.Limit, request.Offset)
	if err != nil {
		return nil, err
	}
	return utils.Map(rows, searchRow.toDomain), nil
}

func (r *CatalogRepo) ConstellationByStar(ctx context.Context, id catalog.ObjectID) (*catalog.Constellation, error) {
	row, err := r.m.CollectOne[constellationIDRow](ctx, `
		select constellation_id as id
		from catalog.constellation_stars
		where star_id = $1
	`, id)
	if errors.Is(err, catalog.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	res, err := r.loadConstellations(ctx, &row.ID)
	if err != nil {
		return nil, err
	}
	return &res[0], nil
}

func (r *CatalogRepo) loadConstellations(ctx context.Context, id *int16) ([]catalog.Constellation, error) {
	consts, err := r.m.CollectRows[constellationRow](ctx, `
		select id, identifier_iau, identifier_common from catalog.constellations
		where $1::smallint is null or id = $1
		order by identifier_common
	`, id)
	if err != nil {
		return nil, err
	}

	bounds, err := r.m.CollectRows[boundaryRow](ctx, `
		select constellation_id, start_ra_rad, start_dec_rad, end_ra_rad, end_dec_rad
		from catalog.constellation_boundary_v
		where $1::smallint is null or constellation_id = $1
		order by constellation_id, seq_no
	`, id)
	if err != nil {
		return nil, err
	}

	pattern, err := r.m.CollectRows[patternRow](ctx, `
		select constellation_id, segment_no, id, ra_rad, dec_rad
		from catalog.constellation_pattern_v
		where $1::smallint is null or constellation_id = $1
		order by constellation_id, segment_no, seq_no
	`, id)
	if err != nil {
		return nil, err
	}

	res := make([]catalog.Constellation, len(consts))
	byID := make(map[int16]*catalog.Constellation, len(consts))
	for i, c := range consts {
		res[i] = catalog.Constellation{ID: c.ID, IAU: c.IAU, Common: c.Common}
		byID[c.ID] = &res[i]
	}

	for _, b := range bounds {
		c := byID[b.ConstellationID]
		c.Boundary = append(c.Boundary, b.toDomain())
	}

	var prev struct{ constellation, segment int16 }
	for _, p := range pattern {
		c := byID[p.ConstellationID]
		if p.ConstellationID != prev.constellation || p.SegmentNo != prev.segment {
			c.Pattern = append(c.Pattern, nil)
			prev.constellation, prev.segment = p.ConstellationID, p.SegmentNo
		}
		last := len(c.Pattern) - 1
		c.Pattern[last] = append(c.Pattern[last], p.toDomain())
	}

	return res, nil
}

type objectRefRow struct {
	ID         int64   `db:"id"`
	Class      string  `db:"class"`
	CommonName *string `db:"common_name"`
}

func (r objectRefRow) toDomain() catalog.ObjectRef {
	return catalog.ObjectRef{ID: catalog.ObjectID(r.ID), Class: catalog.Class(r.Class), CommonName: r.CommonName}
}

type positionRow struct {
	ID  int64   `db:"id"`
	RA  float64 `db:"ra_rad"`
	Dec float64 `db:"dec_rad"`
}

func (r positionRow) toDomain() catalog.ObjectPosition {
	return catalog.ObjectPosition{ID: catalog.ObjectID(r.ID), Position: coordinates.NewEquatorial(r.RA, r.Dec)}
}

type collectionRow struct {
	ID          int64   `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}

func (r collectionRow) toDomain() catalog.Collection {
	return catalog.Collection{ID: catalog.CollectionID(r.ID), Name: r.Name, Description: r.Description}
}

type searchRow struct {
	objectRefRow
	Identifier string `db:"identifier"`
	Collection string `db:"collection"`
}

func (r searchRow) toDomain() catalog.SearchHit {
	return catalog.SearchHit{ObjectRef: r.objectRefRow.toDomain(), Identifier: r.Identifier, Collection: r.Collection}
}

// objectRow: subtype presence is determined by has_* flags, not by NULL checks,
// because all dso_data fields are nullable.
type objectRow struct {
	ID          int64    `db:"id"`
	Class       string   `db:"class"`
	RA          float64  `db:"ra_rad"`
	Dec         float64  `db:"dec_rad"`
	Distance    *float64 `db:"distance"`
	TypeName    string   `db:"type_name"`
	CommonName  *string  `db:"common_name"`
	Description *string  `db:"description"`
	ImageURL    *string  `db:"image_url"`

	HasStar       bool     `db:"has_star"`
	StarMag       *float64 `db:"star_mag"`
	SpectralClass *string  `db:"spectral_class"`

	HasDSO      bool     `db:"has_dso"`
	DSOMag      *float64 `db:"dso_mag"`
	DSOMajor    *float64 `db:"dso_major"`
	DSOMinor    *float64 `db:"dso_minor"`
	DSOPosAngle *float64 `db:"dso_pos_angle"`
}

func (r objectRow) toDomain() (catalog.Object, error) {
	o := catalog.Object{
		ID:       catalog.ObjectID(r.ID),
		Class:    catalog.Class(r.Class),
		Position: coordinates.NewEquatorial(r.RA, r.Dec),
		Distance: r.Distance,
		Metadata: catalog.Metadata{
			TypeName:    r.TypeName,
			CommonName:  r.CommonName,
			Description: r.Description,
			ImageURL:    r.ImageURL,
		},
	}

	// The schema does not guarantee consistency between class and subtype tables,
	// so a mismatch is a data error.
	switch o.Class {
	case catalog.ClassStar:
		if !r.HasStar || r.StarMag == nil {
			return catalog.Object{}, fmt.Errorf("object %d: class star, but no star_data row", r.ID)
		}
		o.Star = &catalog.StarData{VisualMag: *r.StarMag, SpectralClass: r.SpectralClass}
	case catalog.ClassDSO:
		if !r.HasDSO {
			return catalog.Object{}, fmt.Errorf("object %d: class dso, but no dso_data row", r.ID)
		}
		o.DSO = &catalog.DSOData{
			VisualMag: r.DSOMag,
			MajorAxis: r.DSOMajor,
			MinorAxis: r.DSOMinor,
			PosAngle:  r.DSOPosAngle,
		}
	}
	return o, nil
}

type constellationRow struct {
	ID     int16  `db:"id"`
	IAU    string `db:"identifier_iau"`
	Common string `db:"identifier_common"`
}

type constellationIDRow struct {
	ID int16 `db:"id"`
}

type boundaryRow struct {
	ConstellationID int16   `db:"constellation_id"`
	StartRA         float64 `db:"start_ra_rad"`
	StartDec        float64 `db:"start_dec_rad"`
	EndRA           float64 `db:"end_ra_rad"`
	EndDec          float64 `db:"end_dec_rad"`
}

func (r boundaryRow) toDomain() catalog.BoundarySegment {
	return catalog.BoundarySegment{
		Start: coordinates.NewEquatorial(r.StartRA, r.StartDec),
		End:   coordinates.NewEquatorial(r.EndRA, r.EndDec),
	}
}

type patternRow struct {
	positionRow
	ConstellationID int16 `db:"constellation_id"`
	SegmentNo       int16 `db:"segment_no"`
}
