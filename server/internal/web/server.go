package web

import (
	"astrohop/internal/account"
	"astrohop/internal/mission"
	"astrohop/internal/search"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	accountHandler                                                     *account.Handler
	searchHandler                                                      *search.Handler
	missionHandler                                                     *mission.Handler
	authMware, permViewMware, permEditMware, permDeleteMware, errMware gin.HandlerFunc
}

func NewServer(
	accountHandler *account.Handler,
	searchHandler *search.Handler,
	missionHandler *mission.Handler,
	authMware, permViewMware, permEditMware, permDeleteMware, errMware gin.HandlerFunc,
) *Server {
	return &Server{
		accountHandler:  accountHandler,
		searchHandler:   searchHandler,
		missionHandler:  missionHandler,
		authMware:       authMware,
		permViewMware:   permViewMware,
		permEditMware:   permEditMware,
		permDeleteMware: permDeleteMware,
		errMware:        errMware,
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
	r.GET("/api/v1/search/objects", s.searchHandler.HandleSearchObjectsByName())
	r.POST("/api/v1/accounts", s.accountHandler.HandleCreateAccount())

	r.POST("/api/v1/missions", s.authMware, s.missionHandler.HandleCreateMission())
	r.GET("/api/v1/missions", s.authMware, s.missionHandler.HandleGetAccountMissions())
	r.GET("/api/v1/missions/:mission_id", s.authMware, s.permViewMware, s.missionHandler.HandleGetMission())
	r.PATCH("/api/v1/missions/:mission_id", s.authMware, s.permEditMware, s.missionHandler.HandleUpdateMission())
	r.PATCH("/api/v1/missions/:mission_id/visibility", s.authMware, s.permEditMware, s.missionHandler.HandleUpdateMissionVisibility())
	r.DELETE("/api/v1/missions/:mission_id", s.authMware, s.permDeleteMware, s.missionHandler.HandleDeleteMission())
	r.GET("/api/v1/missions/:mission_id/stream", s.authMware, s.permViewMware, s.missionHandler.HandleGetMissionStream())
}
