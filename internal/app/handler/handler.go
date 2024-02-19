package handler

import (
	"compress/gzip"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/logger"
	"github.com/MrSwed/go-musthave-shortener/internal/app/middleware"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	s service.Service
	r *gin.Engine
	c *config.Auth
}

func NewHandler(s service.Service, c *config.Auth) *Handler { return &Handler{s: s, c: c} }

func (h *Handler) Handler() http.Handler {
	h.r = gin.New()
	h.r.Use(logger.Logger())
	h.r.Use(middleware.Compress(gzip.DefaultCompression))
	h.r.Use(middleware.Decompress())
	h.r.Use(h.Auth())

	h.r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})
	rootRoute := h.r.Group("/")
	rootRoute.POST("", h.MakeShort())
	rootRoute.GET("/ping", h.GetDBPing())
	rootRoute.GET("/:id", h.GetShort())

	apiRoute := rootRoute.Group(constant.APIRoute)
	shortAPIRoute := apiRoute.Group(constant.ShortenRoute)
	shortAPIRoute.POST("", h.MakeShortJSON())
	shortAPIRoute.POST(constant.BatchRoute, h.MakeShortBatch())

	userAPIRoute := apiRoute.Group(constant.UserRoute)
	userAPIRoute.GET(constant.URLsRoute, h.GetAllByUser())

	return h.r
}

func (h *Handler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.c == nil {
			logrus.Error("need config with secret key")
			c.Next()
			return
		}

		authStr, err := c.Cookie(constant.CookieAuthName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			logrus.Error("Error get cookie", err)
		}

		astc, nonce, err := h.c.AuthCipher()
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		var (
			user domain.UserInfo
		)

		if authStr != "" {
			authBytes, err := hex.DecodeString(authStr)
			if err != nil {
				logrus.Error("hex decode string error: ", err)
			}

			uuidBytes, err := astc.Open(nil, nonce, authBytes, nil)
			if err != nil {
				logrus.Error("ahead open error: ", err)
			}

			if user, err = h.s.GetUser(c, string(uuidBytes)); err != nil && !errors.Is(err, myErr.ErrNotExist) {
				logrus.Error("get user error", err)
			}
		}

		if user.ID == "" {
			if user.ID, err = h.s.NewUser(c); err != nil {
				logrus.Error(err)
			}

			authBytes := astc.Seal(nil, nonce, []byte(user.ID), nil)
			authStr = hex.EncodeToString(authBytes)
			c.SetCookie(constant.CookieAuthName, authStr, 0, "", strings.Split(c.Request.Host, ":")[0], false, false)
		}

		if user.ID == "" { // or what ?
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(constant.ContextUserValueName, user.ID)
		c.Next()
	}
}
