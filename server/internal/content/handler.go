package content

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

func (h *Handler) HandleGetMissionMap() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionId := c.GetInt64("mission_id")
		mmap, size, err := h.svc.GetMissionMap(c.Request.Context(), missionId)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.DataFromReader(http.StatusOK, size, "image/svg+xml", mmap, map[string]string{
			"Content-Encoding": "gzip",
		})
	}
}
