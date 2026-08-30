package middleware

import (
	"astrohop/pkg/apperr"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthRepo interface {
	ValidateOwnership(ctx context.Context, missionId int64, accessToken string) (bool, error)
}

func NewAuth(
	repo AuthRepo,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		missionIdStr := c.Param("id")
		if missionIdStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}
		missionId, err := strconv.Atoi(missionIdStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
			return
		}
		accessToken := c.GetHeader("X-Access-Token")
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-Access-Token header"})
			return
		}

		valid, err := repo.ValidateOwnership(c.Request.Context(), int64(missionId), accessToken)
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			return
		}

		c.Set("mission_id", missionId)
	}
}
