package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
			h.log.WithField("Error", err).Error("Error reate new short")
		}
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusCreated, html)
	}
}

func (h *Handler) GetShort() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if newURL, err := h.s.GetFromShort(id); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			c.Redirect(http.StatusTemporaryRedirect, newURL)
			return
		}
	}
}
