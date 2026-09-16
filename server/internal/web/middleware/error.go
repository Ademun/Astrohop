package middleware

import (
	"astrohop/pkg/apperr"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

var kindToStatus = map[apperr.Kind]int{
	apperr.KindNotFound:     http.StatusNotFound,
	apperr.KindValidation:   http.StatusUnprocessableEntity,
	apperr.KindForbidden:    http.StatusForbidden,
	apperr.KindUnauthorized: http.StatusUnauthorized,
	apperr.KindConflict:     http.StatusConflict,
	apperr.KindInternal:     http.StatusInternalServerError,
	apperr.KindTimeout:      http.StatusGatewayTimeout,
}

func NewError(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		var appErr *apperr.Err
		kind, code, msg := apperr.KindInternal, "unknown", "internal server error"
		if errors.As(err, &appErr) {
			kind, code, msg = appErr.Kind, appErr.Code, appErr.Msg
		}

		status, ok := kindToStatus[kind]
		if !ok {
			status = http.StatusInternalServerError
		}

		fields := []any{
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", status,
			"code", code,
			"cause", err.Error(),
		}
		if accountID, ok := c.Get("account_id"); ok {
			fields = append(fields, "account_id", accountID)
		}
		if missionID, ok := c.Get("mission_id"); ok {
			fields = append(fields, "mission_id", missionID)
		}

		if status >= http.StatusInternalServerError {
			log.Error("request failed", fields...)
		} else {
			log.Warn("request failed", fields...)
		}

		c.AbortWithStatusJSON(status, gin.H{"error": msg, "code": code})
	}
}
