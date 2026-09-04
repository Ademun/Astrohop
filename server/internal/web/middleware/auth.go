package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthRepo interface {
	GetAccountIDByKey(ctx context.Context, key string) (int64, error)
}

func NewAuth(
	repo AuthRepo,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("Authorization")
		if rawKey == "" || !strings.HasPrefix(rawKey, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is empty"})
			return
		}
		key := strings.TrimPrefix(rawKey, "Bearer ")

		id, err := repo.GetAccountIDByKey(c.Request.Context(), key)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		c.Set("account_id", id)
	}
}
