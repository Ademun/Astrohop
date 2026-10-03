package postgres

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/catalog"
	"astrohop/pkg/db"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type CatalogRepo struct {
	m *db.Manager
}

func NewCatalogRepo(m *db.Manager) *CatalogRepo {
	return &CatalogRepo{m: m}
}

func (r *CatalogRepo) Collections(ctx context.Context) ([]catalog.Collection, error) {
	rows, err := collectRows[collectionRow](ctx, r.m, `
		select id, name, description
		from catalog.collections
		order by name
	`)
	if err != nil {
		return nil, err
	}
	return mapSlice(rows, collectionRow.ToDomain), nil
}

func (r *CatalogRepo) CollectionObjects(ctx context.Context, id catalog.CollectionID, page catalog.PageRequest) (catalog.Page[catalog.CollectionMember], error) {
	count, err := collectOne[totalRow](ctx, r.m, `
		select total
		from catalog.collection_object_counts_v
		where collection_id = $1
	`, id)
	if err != nil {
		return catalog.Page[catalog.CollectionMember]{}, err
	}

	res := catalog.Page[catalog.CollectionMember]{Total: count.Total}
	if count.Total == 0 || page.Offset >= count.Total {
		return res, nil
	}

	// Natural order: alphabetic prefix, then the first number numerically (M2 before M10).
	rows, err := collectRows[collectionMemberRow](ctx, r.m, `
		select object_id as id, class, common_name, identifier
		from catalog.collection_object_details_v
		where collection_id = $1
		order by
			(regexp_match(identifier, '^\D*'))[1],
			coalesce((regexp_match(identifier, '\d+'))[1]::numeric, 0),
			identifier
		limit $2 offset $3
	`, id, page.Limit, page.Offset)
	if err != nil {
		return catalog.Page[catalog.CollectionMember]{}, err
	}

	res.Items = mapSlice(rows, collectionMemberRow.ToDomain)
	return res, nil
}

func (r *CatalogRepo) Object(ctx context.Context, id catalog.ObjectID) (catalog.Object, error) {
	row, err := collectOne[objectRow](ctx, r.m, `
		select * from catalog.object_full_v where id = $1
	`, id)
	if err != nil {
		return catalog.Object{}, err
	}
	return row.ToDomain()
}

func (r *CatalogRepo) Positions(ctx context.Context, ids []catalog.ObjectID) ([]catalog.ObjectPosition, error) {
	rows, err := collectRows[positionRow](ctx, r.m, `
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

	return mapSlice(rows, positionRow.ToDomain), nil
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (r *CatalogRepo) Search(ctx context.Context, query string, limit int) ([]catalog.SearchHit, error) {
	rows, err := collectRows[searchRow](ctx, r.m, `
		with q as (
			select catalog.norm_ident($1) as v, catalog.norm_ident($2) as pat
		)
		select d.object_id as id, d.class, d.common_name, d.identifier, d.collection
		from catalog.collection_object_details_v d, q
		where d.identifier_norm like '%' || q.pat || '%'
		order by
			case
				when d.identifier_norm = q.v then 0
				when d.identifier_norm like q.pat || '%' then 1
				else 2
			end,
			length(d.identifier),
			d.identifier
		limit $3
	`, query, likeEscaper.Replace(query), limit)
	if err != nil {
		return nil, err
	}
	return mapSlice(rows, searchRow.ToDomain), nil
}

func (r *CatalogRepo) ConstellationByStar(ctx context.Context, id catalog.ObjectID) (*catalog.Constellation, error) {
	row, err := collectOne[constellationIDRow](ctx, r.m, `
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
	consts, err := collectRows[constellationRow](ctx, r.m, `
		select id, identifier_iau, identifier_common from catalog.constellations
		where $1::smallint is null or id = $1
		order by identifier_common
	`, id)
	if err != nil {
		return nil, err
	}

	bounds, err := collectRows[boundaryRow](ctx, r.m, `
		select constellation_id, start_ra_rad, start_dec_rad, end_ra_rad, end_dec_rad
		from catalog.constellation_boundary_v
		where $1::smallint is null or constellation_id = $1
		order by constellation_id, seq_no
	`, id)
	if err != nil {
		return nil, err
	}

	pattern, err := collectRows[patternRow](ctx, r.m, `
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
		c.Boundary = append(c.Boundary, b.ToDomain())
	}

	var prev struct{ constellation, segment int16 }
	for _, p := range pattern {
		c := byID[p.ConstellationID]
		if p.ConstellationID != prev.constellation || p.SegmentNo != prev.segment {
			c.Pattern = append(c.Pattern, nil)
			prev.constellation, prev.segment = p.ConstellationID, p.SegmentNo
		}
		last := len(c.Pattern) - 1
		c.Pattern[last] = append(c.Pattern[last], p.ToDomain())
	}

	return res, nil
}

func equatorial(raRad, decRad float64) coordinates.Equatorial {
	return coordinates.NewEquatorial(raRad*coordinates.RadToDeg, decRad*coordinates.RadToDeg)
}

func mapSlice[R, T any](in []R, f func(R) T) []T {
	out := make([]T, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func collectRows[R any](ctx context.Context, m *db.Manager, sql string, args ...any) ([]R, error) {
	rows, err := m.GetExecutor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[R])
}

func collectOne[R any](ctx context.Context, m *db.Manager, sql string, args ...any) (R, error) {
	var zero R
	rows, err := m.GetExecutor(ctx).Query(ctx, sql, args...)
	if err != nil {
		return zero, err
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[R])
	if errors.Is(err, pgx.ErrNoRows) {
		return zero, catalog.ErrNotFound
	}
	return row, err
}

type objectRefRow struct {
	ID         int64   `db:"id"`
	Class      string  `db:"class"`
	CommonName *string `db:"common_name"`
}

func (r objectRefRow) ToDomain() catalog.ObjectRef {
	return catalog.ObjectRef{ID: catalog.ObjectID(r.ID), Class: catalog.Class(r.Class), CommonName: r.CommonName}
}

type positionRow struct {
	ID  int64   `db:"id"`
	RA  float64 `db:"ra_rad"`
	Dec float64 `db:"dec_rad"`
}

func (r positionRow) ToDomain() catalog.ObjectPosition {
	return catalog.ObjectPosition{ID: catalog.ObjectID(r.ID), Position: equatorial(r.RA, r.Dec)}
}

type collectionRow struct {
	ID          int64   `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}

func (r collectionRow) ToDomain() catalog.Collection {
	return catalog.Collection{ID: catalog.CollectionID(r.ID), Name: r.Name, Description: r.Description}
}

type totalRow struct {
	Total int `db:"total"`
}

type collectionMemberRow struct {
	objectRefRow
	Identifier string `db:"identifier"`
}

func (r collectionMemberRow) ToDomain() catalog.CollectionMember {
	return catalog.CollectionMember{ObjectRef: r.objectRefRow.ToDomain(), Identifier: r.Identifier}
}

type searchRow struct {
	objectRefRow
	Identifier string `db:"identifier"`
	Collection string `db:"collection"`
}

func (r searchRow) ToDomain() catalog.SearchHit {
	return catalog.SearchHit{ObjectRef: r.objectRefRow.ToDomain(), Identifier: r.Identifier, Collection: r.Collection}
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

func (r objectRow) ToDomain() (catalog.Object, error) {
	o := catalog.Object{
		ID:       catalog.ObjectID(r.ID),
		Class:    catalog.Class(r.Class),
		Position: equatorial(r.RA, r.Dec),
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

func (r boundaryRow) ToDomain() catalog.BoundarySegment {
	return catalog.BoundarySegment{
		Start: equatorial(r.StartRA, r.StartDec),
		End:   equatorial(r.EndRA, r.EndDec),
	}
}

type patternRow struct {
	positionRow
	ConstellationID int16 `db:"constellation_id"`
	SegmentNo       int16 `db:"segment_no"`
}
