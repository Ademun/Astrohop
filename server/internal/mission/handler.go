package mission

import (
	"astrohop/pkg/apperr"
	"astrohop/pkg/logger"
	"io"
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreateMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		var json DataDTO
		if err := c.ShouldBindJSON(&json); err != nil {
			apperr.HandleHttp(c, apperr.New(http.StatusUnprocessableEntity, "invalid mission", err))
			return
		}

		id, err := h.svc.CreateMission(c.Request.Context(), json.ToDomain(), accountID)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"mission_id": id})
	}
}

func (h *Handler) HandleUpdateMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionID := uuid.MustParse(c.GetString("mission_id"))
		var json DataDTO
		if err := c.ShouldBindJSON(&json); err != nil {
			apperr.HandleHttp(c, apperr.New(http.StatusUnprocessableEntity, "invalid mission", err))
			return
		}
		if err := h.svc.UpdateMissionData(c.Request.Context(), missionID, json.ToDomain()); err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (h *Handler) HandleUpdateMissionVisibility() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionID := uuid.MustParse(c.GetString("mission_id"))
		isPublic := c.Query("public") == "true"
		if err := h.svc.UpdateMissionVisibility(c.Request.Context(), missionID, isPublic); err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (h *Handler) HandleDeleteMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionID := uuid.MustParse(c.GetString("mission_id"))
		if err := h.svc.DeleteMission(c.Request.Context(), missionID); err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (h *Handler) HandleGetMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionID := uuid.MustParse(c.GetString("mission_id"))
		mission, err := h.svc.GetMission(c.Request.Context(), missionID)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.JSON(http.StatusOK, mission.ToDTO())
	}
}

func (h *Handler) HandleGetMissionStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		missionID := uuid.MustParse(c.GetString("mission_id"))
		stream, err := h.svc.GetMissionStream(c.Request.Context(), missionID)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Stream(func(w io.Writer) bool {
			select {
			case <-c.Request.Context().Done():
				return false
			case progress, ok := <-stream:
				if !ok {
					return false
				}
				msg := gin.H{
					"progress": progress.Progress,
				}
				if progress.Error != nil {
					logger.L().Error(progress.Error)
					msg["error"] = progress.Error.Error()
					c.SSEvent("message", msg)
					return false
				}
				if progress.Payload != nil {
					msg["payload"] = progress.Payload
					c.SSEvent("message", msg)
				}
				return true
			}
		})
	}
}
