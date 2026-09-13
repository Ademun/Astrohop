package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
			c.Error(err)
			return
		}
		c.Header("Authorization", "Bearer "+key)
		c.Status(http.StatusCreated)
	}
}

func (h *Handler) HandleDevicePairing() gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(err)
			return
		}
		defer conn.Close()

		requestID, err := h.svc.CreatePairingRequest(c.Request.Context())
		if err != nil {
			c.Error(err)
			return
		}

		if err := conn.WriteJSON(gin.H{
			"requestID": requestID,
		}); err != nil {
			c.Error(err)
			return
		}

	}
}
