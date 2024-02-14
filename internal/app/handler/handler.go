package handler

import (
	"compress/gzip"
	"net/http"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/logger"
	"github.com/MrSwed/go-musthave-shortener/internal/app/middleware"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	s   service.Service
	r   *gin.Engine
	log *logrus.Logger
}

func NewHandler(s service.Service) *Handler { return &Handler{s: s} }

func (h *Handler) Handler() http.Handler {
	h.r = gin.New()
	h.r.Use(logger.Logger())
	h.r.Use(middleware.Compress(gzip.DefaultCompression, h.log))
	h.r.Use(middleware.Decompress(h.log))

	h.r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})
	rootRoute := h.r.Group("/")
	rootRoute.POST("", h.MakeShort())
	rootRoute.GET("/ping", h.GetDBPing())
	rootRoute.GET("/:id", h.GetShort())

	apiRoute := rootRoute.Group(config.APIRoute)
	shortAPIRoute := apiRoute.Group(config.ShortenRoute)
	shortAPIRoute.POST("", h.MakeShortJSON())
	shortAPIRoute.POST(config.BatchRoute, h.MakeShortBatch())

	return h.r
}
