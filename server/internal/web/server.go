package web

import (
	"astrohop/internal/account"
	"astrohop/internal/catalog"
	"astrohop/internal/mission"
	"astrohop/internal/pairing"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	accountHandler      *account.Handler
	catalogHandler      *catalog.Handler
	missionHandler      *mission.Handler
	pairingHandler      *pairing.Handler
	authMware, errMware gin.HandlerFunc
}

func NewServer(
	accountHandler *account.Handler,
	catalogHandler *catalog.Handler,
	missionHandler *mission.Handler,
	pairingHandler *pairing.Handler,
	authMware, errMware gin.HandlerFunc,
) *Server {
	return &Server{
		accountHandler: accountHandler,
		catalogHandler: catalogHandler,
		missionHandler: missionHandler,
		pairingHandler: pairingHandler,
		authMware:      authMware,
		errMware:       errMware,
	}
}

func (s *Server) Server(addr string) *http.Server {
	r := gin.Default()
	r.Use(s.errMware)
	s.setupCors(r)
	s.registerRoutes(r)
	return &http.Server{
		Addr:    addr,
		Handler: r.Handler(),
	}
}

func (s *Server) setupCors(r *gin.Engine) {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"http://localhost:5173", "http://192.168.1.109:5173"}
	cfg.AddExposeHeaders("Authorization")
	cfg.AddAllowHeaders("Authorization")
	r.Use(cors.New(cfg))
}

func (s *Server) registerRoutes(r *gin.Engine) {
	r.GET("/api/v1/catalog/collections", s.catalogHandler.HandleCollections())
	r.GET("/api/v1/catalog/collections/:id/objects", s.catalogHandler.HandleCollectionObjects())
	r.GET("/api/v1/catalog/objects/:id", s.catalogHandler.HandleObject())
	r.GET("/api/v1/catalog/objects/search", s.catalogHandler.HandleSearch())

	r.POST("/api/v1/accounts", s.accountHandler.HandleCreateAccount())

	r.POST("/api/v1/missions", s.authMware, s.missionHandler.HandleCreateMission())
	r.GET("/api/v1/missions/:mission_id", s.authMware, s.missionHandler.HandleGetMission())
	r.GET("/api/v1/missions", s.authMware, s.missionHandler.HandleGetAccountMissions())
	r.GET("/api/v1/missions/:mission_id/stream", s.authMware, s.missionHandler.HandleGetMissionStream())
	r.PATCH("/api/v1/missions/:mission_id/info", s.authMware, s.missionHandler.HandleUpdateMissionInformation())
	r.PATCH("/api/v1/missions/:mission_id/data", s.authMware, s.missionHandler.HandleUpdateMissionData())
	r.DELETE("/api/v1/missions/:mission_id", s.authMware, s.missionHandler.HandleDeleteMission())
	r.GET("/api/v1/pairing", s.pairingHandler.HandleStartPairing())
	r.GET("/api/v1/pairing/:req_id", s.pairingHandler.HandleAttachTarget())
}
