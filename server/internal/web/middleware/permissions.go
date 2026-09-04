package middleware

import (
	"astrohop/internal/account"
	"context"
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"
)

type Action string

const (
	ActionView   Action = "view"
	ActionEdit   Action = "edit"
	ActionDelete Action = "delete"
)

type PermissionsRepo interface {
	GetMissionPermissions(ctx context.Context, accountID int64, missionID uuid.UUID) (*account.MissionPermissions, error)
}

func NewPermissions(repo PermissionsRepo, action Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.GetInt64("account_id")
		missionIDS := c.Param("mission_id")
		if missionIDS == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing mission_id"})
			return
		}
		missionID, err := uuid.Parse(missionIDS)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid mission_id"})
			return
		}

		permissions, err := repo.GetMissionPermissions(c.Request.Context(), accountID, missionID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if permissions.IsOwner {
			c.Set("mission_id", missionID.String())
			return
		}

		if !permissions.IsMissionPublic {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if action == ActionView {
			c.Set("mission_id", missionID.String())
			return
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}
