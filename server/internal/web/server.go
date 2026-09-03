package web

import (
	"astrohop/internal/account"
	"astrohop/internal/infrastructure/postgres"
	"astrohop/internal/mission"
	"astrohop/internal/search"
	"astrohop/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	missionRepo    *postgres.MissionRepo
	accountRepo    *postgres.AccountRepo
	accountHandler *account.Handler
	searchHandler  *search.Handler
	missionHandler *mission.Handler
}

func NewServer(
	missionRepo *postgres.MissionRepo,
	accountRepo *postgres.AccountRepo,
	accountHandler *account.Handler,
	searchHandler *search.Handler,
	missionHandler *mission.Handler,
) *Server {
	return &Server{
		missionRepo:    missionRepo,
		accountRepo:    accountRepo,
		accountHandler: accountHandler,
		searchHandler:  searchHandler,
		missionHandler: missionHandler,
	}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	s.setupCors(r)
	s.registerRoutes(r)
	return r.Run(addr)
}

func (s *Server) setupCors(r *gin.Engine) {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"http://localhost:5173"}
	cfg.AddExposeHeaders("Authorization")
	cfg.AddAllowHeaders("Authorization")
	r.Use(cors.New(cfg))
}

func (s *Server) registerRoutes(r *gin.Engine) {
	authM := middleware.NewAuth(s.accountRepo)
	permViewM := middleware.NewPermissions(s.missionRepo, middleware.ActionView)
	permEditM := middleware.NewPermissions(s.missionRepo, middleware.ActionEdit)
	permDeleteM := middleware.NewPermissions(s.missionRepo, middleware.ActionDelete)

	r.GET("/api/v1/search/objects", s.searchHandler.HandleSearchObjectsByName())
	r.POST("/api/v1/accounts", s.accountHandler.HandleCreateAccount())

	r.POST("/api/v1/missions", authM, s.missionHandler.HandleCreateMission())
	r.GET("/api/v1/missions/:mission_id", authM, permViewM, s.missionHandler.HandleGetMission())
	r.PATCH("/api/v1/missions/:mission_id", authM, permEditM, s.missionHandler.HandleUpdateMission())
	r.DELETE("/api/v1/missions/:mission_id", authM, permDeleteM, s.missionHandler.HandleDeleteMission())
	r.GET("/api/v1/missions/:mission_id/stream", authM, permViewM, s.missionHandler.HandleGetMissionStream())
}
