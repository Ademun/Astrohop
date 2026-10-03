package catalog

import (
	"astrohop/pkg/apperr"
	"astrohop/pkg/utils"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	maxSearchQueryRuneCount = 50
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		_ = c.Error(apperr.Validation(ErrInvalidID, "id must be a positive integer", nil))
		return 0, false
	}
	return id, true
}

func (h *Handler) HandleCollections() gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h.svc.Collections(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, utils.Map(res, newCollectionDTO))
	}
}

func (h *Handler) HandleCollectionObjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}

		var q CollectionObjectsQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			_ = c.Error(apperr.Validation(ErrInvalidPage, "limit and offset must be integers", nil))
			return
		}

		res, err := h.svc.CollectionObjects(c.Request.Context(), CollectionID(id), q.ToDomain())
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, newCollectionObjectsPageDTO(res))
	}
}

func (h *Handler) HandleObject() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}

		res, err := h.svc.Object(c.Request.Context(), ObjectID(id))
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, newObjectDTO(res))
	}
}

func (h *Handler) HandleSearch() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		if utf8.RuneCountInString(q) > maxSearchQueryRuneCount {
			_ = c.Error(apperr.Validation(ErrSearchQueryTooLong, "search query is too long", nil))
			return
		}

		res, err := h.svc.Search(c.Request.Context(), q)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, utils.Map(res, newSearchHitDTO))
	}
}
