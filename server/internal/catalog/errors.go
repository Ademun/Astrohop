package catalog

import "errors"

var ErrNotFound = errors.New("catalog: not found")

const (
	ErrListCollections  = "catalog.list_collections"
	ErrInvalidPage      = "catalog.invalid_page"
	ErrGetObject        = "catalog.get_object"
	ErrObjectNotFound   = "catalog.object_not_found"
	ErrGetPositions     = "catalog.get_positions"
	ErrSearch           = "catalog.search"
	ErrConstellationsOf = "catalog.constellations_of"
	ErrInvalidID        = "catalog.invalid_id"
)
