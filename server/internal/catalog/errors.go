package catalog

import "errors"

var ErrNotFound = errors.New("catalog: not found")

const (
	ErrListCollections    = "catalog.list_collections"
	ErrCollectionObjects  = "catalog.collection_objects"
	ErrCollectionNotFound = "catalog.collection_not_found"
	ErrInvalidPage        = "catalog.invalid_page"
	ErrGetObject          = "catalog.get_object"
	ErrObjectNotFound     = "catalog.object_not_found"
	ErrGetPositions       = "catalog.get_positions"
	ErrSearch             = "catalog.search"
	ErrSearchEmptyQuery   = "catalog.search_empty_query"
	ErrConstellationGuide = "catalog.constellation_guide"
	ErrConstellations     = "catalog.constellations"
	ErrConstellationsOf   = "catalog.constellations_of"
	ErrInvalidID          = "catalog.invalid_id"
	ErrSearchQueryTooLong = "catalog.search_query_too_long"
)
