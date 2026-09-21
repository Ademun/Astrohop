package mission

import (
	"astrohop/pkg/apperr"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			_ = c.Error(apperr.Validation(ErrValidation, "invalid mission object", nil))
			return
		}

		id, err := h.svc.CreateMission(c.Request.Context(), json.ToDomain(), accountID)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"mission_id": id})
	}
}

func (h *Handler) HandleGetMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionID, err := h.getMissionID(c)
		if err != nil {
			_ = c.Error(err)
			return
		}

		mission, err := h.svc.GetMission(c.Request.Context(), missionID, accountID)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, NewMissionFullDTO(mission))
	}
}

func (h *Handler) HandleGetAccountMissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")

		missions, err := h.svc.GetAccountMissions(c.Request.Context(), accountID)
		if err != nil {
			_ = c.Error(err)
			return
		}

		json := make([]MissionShortDTO, len(missions))
		for i, m := range missions {
			json[i] = NewMissionShortDTO(&m)
		}

		c.JSON(http.StatusOK, json)
	}
}

func (h *Handler) HandleGetMissionStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionID, err := h.getMissionID(c)
		if err != nil {
			_ = c.Error(err)
			return
		}
		stream, cancel, err := h.svc.GetMissionStream(c.Request.Context(), missionID, accountID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		defer cancel()

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
					_ = c.Error(progress.Error)
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

func (h *Handler) HandleUpdateMissionInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionID, err := h.getMissionID(c)
		if err != nil {
			_ = c.Error(err)
			return
		}

		var json InformationDTO
		if err := c.ShouldBindJSON(&json); err != nil {
			_ = c.Error(apperr.Validation(ErrValidation, "invalid information object", nil))
			return
		}
		d := json.ToDomain()

		if err := h.svc.UpdateMissionInformation(c.Request.Context(), &d, missionID, accountID); err != nil {
			_ = c.Error(err)
			return
		}
	}
}

func (h *Handler) HandleUpdateMissionData() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionID, err := h.getMissionID(c)
		if err != nil {
			_ = c.Error(err)
			return
		}

		var json DataDTO
		if err := c.ShouldBindJSON(&json); err != nil {
			_ = c.Error(apperr.Validation(ErrValidation, "invalid information object", nil))
			return
		}

		if err := h.svc.UpdateMissionData(c.Request.Context(), json.ToDomain(), missionID, accountID); err != nil {
			_ = c.Error(err)
			return
		}
	}
}

func (h *Handler) HandleDeleteMission() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionID, err := h.getMissionID(c)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if err := h.svc.DeleteMission(c.Request.Context(), missionID, accountID); err != nil {
			_ = c.Error(err)
			return
		}
	}
}

func (h *Handler) getMissionID(c *gin.Context) (uuid.UUID, error) {
	rawID, ok := c.Params.Get("mission_id")
	if !ok {
		return uuid.Nil, apperr.Validation(ErrValidation, "mission_id is required", nil)
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, apperr.Validation(ErrValidation, "mission_id is invalid", nil)
	}
	return id, nil
}
