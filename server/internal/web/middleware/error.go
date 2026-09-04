package middleware

import (
	"astrohop/pkg/apperr"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewError(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		log.Error(err.Error())
		if e, ok := errors.AsType[*apperr.Err](err); ok {
			c.AbortWithStatusJSON(e.Code, gin.H{"error": e.Msg})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
