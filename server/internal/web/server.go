package web

import (
	"astrohop/internal/content"
	"astrohop/internal/infrastructure/postgres"
	"astrohop/internal/planner"
	"astrohop/internal/search"
	"astrohop/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	missionRepo    *postgres.MissionRepo
	searchHandler  *search.Handler
	plannerHandler *planner.Handler
	contentHandler *content.Handler
}

func NewServer(
	missionRepo *postgres.MissionRepo,
	searchHandler *search.Handler,
	plannerHandler *planner.Handler,
	contentHandler *content.Handler,
) *Server {
	return &Server{
		missionRepo:    missionRepo,
		searchHandler:  searchHandler,
		plannerHandler: plannerHandler,
		contentHandler: contentHandler,
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
	cfg.AddExposeHeaders("X-Access-Token")
	cfg.AddAllowHeaders("X-Access-Token")
	r.Use(cors.New(cfg))
}

func (s *Server) registerRoutes(r *gin.Engine) {
	authM := middleware.NewAuth(s.missionRepo)

	r.GET("/api/v1/search/objects", s.searchHandler.HandleSearchObjectsByName())
	r.POST("/api/v1/missions", s.plannerHandler.HandlePreviewMission())

	r.POST("/api/v1/missions/:id/process", authM, s.plannerHandler.HandleProcessMission())
	r.GET("/api/v1/missions/:id/map", authM, s.contentHandler.HandleGetMissionMap())
}
