package account

import (
	"astrohop/pkg/apperr"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreateAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := h.svc.CreateAccount(c.Request.Context())
		if err != nil {
			apperr.HandleHttp(c, err)
			return
		}
		c.Header("Authorization", "Bearer "+key)
		c.Status(http.StatusCreated)
	}
}
