package middleware

import (
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
	ValidateOwnership(ctx context.Context, missionID uuid.UUID, accountID int64) (bool, error)
	IsPublic(ctx context.Context, missionID uuid.UUID) (bool, error)
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

		isOwner, err := repo.ValidateOwnership(c.Request.Context(), missionID, accountID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if isOwner {
			c.Set("mission_id", missionID.String())
			return
		}

		isPublic, err := repo.IsPublic(c.Request.Context(), missionID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		if !isPublic {
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
