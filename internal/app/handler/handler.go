package handler

import (
	"log"
	"net/http"

	"github.com/MrSwed/go-musthave-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	s *service.Service
	r *gin.Engine
}

func NewHandler(s *service.Service) *Handler { return &Handler{s: s} }

func (h *Handler) InitRoutes() *Handler {
	h.r = gin.New()

	h.r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})
	rootRoute := h.r.Group("/")
	rootRoute.POST("/", h.MakeShort())
	rootRoute.GET("/:id", h.GetShort())
	return h
}

func (h *Handler) RunServer(addr string) {
	if h.r == nil {
		h.InitRoutes()
	}
	log.Fatal(http.ListenAndServe(addr, h.r))
}
