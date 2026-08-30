package web

import (
	"astrohop/internal/content"
	"astrohop/internal/planner"
	"astrohop/internal/search"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	searchHandler  *search.Handler
	plannerHandler *planner.Handler
	contentHandler *content.Handler
}

func NewServer(
	searchHandler *search.Handler,
	plannerHandler *planner.Handler,
	contentHandler *content.Handler,
) *Server {
	return &Server{
		searchHandler:  searchHandler,
		plannerHandler: plannerHandler,
		contentHandler: contentHandler,
	}
}

func (s *Server) Start(addr string) error {
	r := gin.Default()
	r.Use(cors.Default())
	s.registerRoutes(r)
	return r.Run(addr)
}

func (s *Server) registerRoutes(r *gin.Engine) {
	r.GET("/api/v1/search/objects", s.searchHandler.HandleSearchObjectsByName())
	r.POST("/api/v1/missions/", s.plannerHandler.HandlePreviewMission())
	r.POST("/api/v1/missions/:id/process", s.plannerHandler.HandleProcessMission())
	r.GET("/api/v1/missions/:id/map", s.contentHandler.HandleGetMissionMap())
}
