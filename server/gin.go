package server

import (
	"github.com/gin-contrib/cors"
	"net/http"
	"time"

	ginprometheus "github.com/mcuadros/go-gin-prometheus"

	"github.com/gin-gonic/gin/binding"

	"github.com/francoispqt/onelog"

	"github.com/gin-gonic/gin"
)

type ConfigServer struct {
	Address string        `default:":8080"`
	Timeout time.Duration `default:"15m"`
}

type ServerHTTP struct {
	cfg    *ConfigServer
	Router gin.IRoutes
	V1     gin.IRoutes
}

func NewServerHTTP(cfg *ConfigServer, logger *onelog.Logger) *ServerHTTP {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	// Меняем валидатор с 8 версий на 10 версию
	binding.Validator = new(defaultValidator)

	v1 := router.Group("/api/v1/")
	v1.Use(Logger(logger))
	v1.Use(gin.Recovery())
	v1.Use(cors.Default())

	return &ServerHTTP{
		cfg:    cfg,
		Router: router,
		V1:     v1,
	}
}

func (s *ServerHTTP) Server() *http.Server {
	return &http.Server{
		ReadTimeout:       s.cfg.Timeout,
		ReadHeaderTimeout: s.cfg.Timeout,
		WriteTimeout:      s.cfg.Timeout,
		IdleTimeout:       s.cfg.Timeout,
		MaxHeaderBytes:    1 << 13, // 8Kb

		Addr:    s.cfg.Address,
		Handler: s.Router.(http.Handler),
	}
}
