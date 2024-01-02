package handler

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"log"
	"net/http"

	"github.com/MrSwed/go-musthave-shortener/internal/app/service"
)

type Handler struct {
	s *service.Service
	r *http.ServeMux
}

func NewHandler(s *service.Service) *Handler { return &Handler{s: s} }

func (h *Handler) InitRoutes() *Handler {
	h.r = http.NewServeMux()
	//h.r.HandleFunc(config.MakeShortRoute, h.MakeShort()
	h.r.HandleFunc(config.MakeShortRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.MakeShort()(w, r)
		case http.MethodGet:
			h.GetShort()(w, r)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return h
}

func (h *Handler) RunServer(addr string) {
	if h.r == nil {
		h.InitRoutes()
	}
	log.Fatal(http.ListenAndServe(addr, h.r))
}
