package planner

import (
	"astrohop/pkg/apperr"
	"astrohop/pkg/logger"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandlePreviewMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		var json MissionInfo
		if err := c.ShouldBindJSON(&json); err != nil {
			apperr.HandleHttp(c, apperr.New(http.StatusUnprocessableEntity, "invalid mission", err))
			return
		}

		result, token, err := h.svc.PreviewMission(c.Request.Context(), json)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Header("X-Access-Token", token)
		c.JSON(http.StatusOK, result)
	}
}

func (h *Handler) HandleProcessMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionId := c.GetInt64("mission_id")
		result, err := h.svc.ProcessMission(missionId)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		c.Stream(func(w io.Writer) bool {
			r, ok := <-result
			if !ok {
				return false
			}
			if r.Error != nil {
				logger.L().Error(r.Error)
				return false
			}

			eventData := gin.H{
				"progress": r.Progress,
			}
			if len(r.Payload) > 0 {
				eventData["payload"] = base64.StdEncoding.EncodeToString(r.Payload)
			}

			c.SSEvent("message", eventData)
			return true
		})
	}
}
