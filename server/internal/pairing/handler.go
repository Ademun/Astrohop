package pairing

import (
	"net/http"

	"astrohop/pkg/apperr"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var allowedOrigins = map[string]struct{}{
	"http://localhost:5173":     {},
	"http://192.168.1.109:5173": {},
}

type Handler struct {
	svc      *Service
	upgrader websocket.Upgrader
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				_, ok := allowedOrigins[r.Header.Get("Origin")]
				return ok
			},
		},
	}
}

func (h *Handler) HandleStartPairing() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		if err := h.svc.StartPairing(c.Request.Context(), conn); err != nil {
			closeWithError(conn, err)
		}
	}
}

func (h *Handler) HandleAttachTarget() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("req_id")

		conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		_, key, err := conn.ReadMessage()
		if err != nil {
			closeWithError(conn, apperr.New(http.StatusBadRequest, "pairing.handler failed to read encryption key", err))
			return
		}

		if err := h.svc.AttachTarget(c.Request.Context(), conn, sessionID, string(key)); err != nil {
			closeWithError(conn, err)
		}
	}
}

func closeWithError(conn *websocket.Conn, err error) {
	reason := err.Error()
	if len(reason) > 123 {
		reason = reason[:123]
	}
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, reason))
	_ = conn.Close()
}
