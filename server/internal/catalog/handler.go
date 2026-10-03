package catalog

import (
	"astrohop/pkg/apperr"
	"astrohop/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		var r SearchRequestDTO
		if err := c.ShouldBindQuery(&r); err != nil {
			_ = c.Error(apperr.Validation(ErrInvalidPage, "limit and offset must be integers", nil))
			return
		}

		if r.Query == nil && r.Collection == nil {
			_ = c.Error(apperr.Validation(ErrInvalidPage, "at least one parameter of query of collection should be not null", nil))
			return
		}

		res, err := h.svc.Search(c.Request.Context(), r.ToDomain())
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, utils.Map(res, newSearchHitDTO))
	}
}
