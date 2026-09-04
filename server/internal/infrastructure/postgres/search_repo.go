package postgres

import (
	"astrohop/internal/search"
	"astrohop/pkg/db"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type SearchRepo struct {
	m *db.Manager
}

func NewSearchRepo(m *db.Manager) *SearchRepo {
	return &SearchRepo{m: m}
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
order by position($1, oids.norm_id),
         strict_word_similarity($1, oids.norm_id) desc
limit 20;
`, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []search.Object{}, nil
		}
		return nil, err
	}

	objects, err := pgx.CollectRows(rows, pgx.RowToStructByName[search.Object])
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (r *SearchRepo) GetObjectsNavData(ctx context.Context, oid []int64) ([]search.NavData, error) {
	rows, err := r.m.GetExecutor(ctx).Query(ctx, `select ra, dec, apparent_mag from data.nav_data where oid = $1`, oid)
	if err != nil {
		return nil, err
	}

	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[search.NavData])
	if err != nil {
		return nil, err
	}

	return data, nil
}
