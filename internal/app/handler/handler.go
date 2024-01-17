package handler

import (
	"net/http"

	"github.com/MrSwed/go-musthave-shortener/internal/app/logger"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	s   service.Service
	r   *gin.Engine
	log *logrus.Logger
}

func NewHandler(s service.Service, log *logrus.Logger) *Handler { return &Handler{s: s, log: log} }

func (h *Handler) InitRoutes() *Handler {
	h.r = gin.New()
	h.r.Use(logger.Logger(h.log))

	h.r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})
	rootRoute := h.r.Group("/")
	rootRoute.POST("/", h.MakeShort())
	rootRoute.GET("/:id", h.GetShort())
	return h
}

func (h *Handler) RunServer(addr string) error {
	if h.r == nil {
		h.InitRoutes()
	}

	return http.ListenAndServe(addr, h.r)
}
