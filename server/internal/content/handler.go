package content

import (
	"astrohop/pkg/apperr"
	"errors"
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

func (h *Handler) HandleGetMissionMap() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionIDStr := c.Param("id")
		missionID, err := strconv.Atoi(missionIDStr)
		if err != nil {
			apperr.HandleHttp(c,
				apperr.New(
					http.StatusUnprocessableEntity,
					"invalid mission id param",
					errors.New("invalid mission id param"),
				),
			)
			return
		}

		mmap, size, err := h.svc.GetMissionMap(c.Request.Context(), int64(missionID))
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}

		c.DataFromReader(http.StatusOK, size, "image/svg+xml", mmap, map[string]string{
			"Content-Encoding": "gzip",
		})
	}
}
