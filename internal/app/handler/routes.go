package handler

import (
	"errors"
	"net/http"

	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/ffjson/ffjson"
)

func (h *Handler) MakeShort() func(c *gin.Context) {
	return func(c *gin.Context) {
		url, err := c.GetRawData()
		if len(url) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error get body")
			return
		}
		html, err := h.s.NewShort(string(url))
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error create new short")
		}
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusCreated, html)
	}
}

func (h *Handler) MakeShortJSON() func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			url    domain.CreateURL
			result domain.ResultURL
			err    error
			body   []byte
		)

		if body, err = c.GetRawData(); err != nil || len(body) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err = ffjson.NewDecoder().Decode(body, &url); err != nil || url.URL == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		result.Result, err = h.s.NewShort(url.URL)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error create new short")
		}
		c.JSON(http.StatusCreated, result)
	}
}

func (h *Handler) GetShort() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if newURL, err := h.s.GetFromShort(id); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				c.AbortWithStatus(http.StatusBadRequest)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
				h.log.WithField("Error", err).Error("Error get new short")
			}
			return
		} else {
			c.Redirect(http.StatusTemporaryRedirect, newURL)
			return
		}
	}
}

func (h *Handler) GetDBPing() func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := h.s.CheckDB(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error get new short")
			return
		} else {
			c.String(http.StatusOK, "Status: ok")
		}
	}
}
