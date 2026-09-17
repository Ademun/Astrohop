package postgres

import (
	"astrohop/internal/search"
	"astrohop/pkg/db"
	"context"

	"github.com/jackc/pgx/v5"
)

type SearchRepo struct {
	m *db.Manager
}

func NewSearchRepo(m *db.Manager) *SearchRepo {
	return &SearchRepo{m: m}
}

type searchObject struct {
	OID        int64  `db:"oid"`
	Identifier string `db:"identifier"`
	ObjectType string `db:"object_type"`
}

func (r searchObject) ToDomain() search.Object {
	return search.Object{
		OID:        r.OID,
		Identifier: r.Identifier,
		ObjectType: r.ObjectType,
	}
}

func (r searchObject) FromDomain(o search.Object) {
	r.OID = o.OID
	r.Identifier = o.Identifier
	r.ObjectType = o.ObjectType
}

type searchNavData struct {
	RA                float64 `db:"ra"`
	Dec               float64 `db:"dec"`
	ApparentMagnitude float32 `db:"apparent_mag"`
}

func (r searchNavData) ToDomain() search.NavData {
	return search.NavData{
		RA:                r.RA,
		Dec:               r.Dec,
		ApparentMagnitude: r.ApparentMagnitude,
	}
}

func (r searchNavData) FromDomain(n search.NavData) {
	r.RA = n.RA
	r.Dec = n.Dec
	r.ApparentMagnitude = n.ApparentMagnitude
}

func (r *SearchRepo) SearchObjectsByName(ctx context.Context, name string) ([]search.Object, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
with prepared_ids as (
    select oid, 
           identifier, 
           lower(replace(identifier, ' ', '')) as norm_id
    from data.object_identifiers
)
select oi.oid,
       oids.identifier,
       oi.object_type
from prepared_ids oids
         inner join data.object_index oi on oi.oid = oids.oid
where oids.norm_id like '%' || $1 || '%'
order by position($1 in oids.norm_id),
         strict_word_similarity($1, oids.norm_id) desc
limit 20;
`, name)
	if err != nil {
		return nil, err
	}

	rowObjects, err := pgx.CollectRows(rows, pgx.RowToStructByName[searchObject])
	if err != nil {
		return nil, err
	}

	objects := make([]search.Object, len(rowObjects))
	for i := range rowObjects {
		objects[i] = rowObjects[i].ToDomain()
	}

	return objects, nil
}

func (r *SearchRepo) GetObjectsNavData(ctx context.Context, oid []int64) ([]search.NavData, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `
select long(pos) * 180.0 / pi() as ra,
       lat(pos) * 180.0 / pi() as dec,
       apparent_mag
from data.nav_data
where oid = any($1)
order by array_position($1, oid)
`, oid)
	if err != nil {
		return nil, err
	}

	rowData, err := pgx.CollectRows(rows, pgx.RowToStructByName[searchNavData])
	if err != nil {
		return nil, err
	}

	data := make([]search.NavData, len(rowData))
	for i := range rowData {
		data[i] = rowData[i].ToDomain()
	}

	return data, nil
}
