package handler

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (h *Handler) MakeShort() func(c *gin.Context) {
	return func(c *gin.Context) {
		url, err := c.GetRawData()
		if len(url) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			log.Printf("No body ")
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Printf("Error get body %s", err)
			return
		}
		html, err := h.s.NewShort(string(url))
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Printf("Error reate new short %s", err)
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
