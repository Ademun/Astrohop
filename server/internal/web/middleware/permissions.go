package middleware

import (
	"astrohop/pkg/apperr"
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

func NewPermissions(
	repo PermissionsRepo,
	action Action,
) gin.HandlerFunc {
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
			apperr.HandleHttp(c, err)
			return
		}
		isPublic, err := repo.IsPublic(c.Request.Context(), missionID)
		if err != nil {
			apperr.HandleHttp(c, err)
		}

		if !isOwner && !isPublic {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if !isOwner && action != ActionView {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("mission_id", missionID.String())
	}
}
