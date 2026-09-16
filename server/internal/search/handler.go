package search

import (
	"astrohop/pkg/apperr"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleSearchObjectsByName() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.Error(apperr.Internal(ErrValidation, "missing object name query param", nil))
			return
		}

		objects, err := h.svc.SearchObjectsByName(c.Request.Context(), name)
		if err != nil {
			c.Error(err)
			return
		}

		dto := make([]ObjectDTO, len(objects))
		for i, object := range objects {
			dto[i] = object.ToDTO()
		}

		c.JSON(http.StatusOK, dto)
	}
}
