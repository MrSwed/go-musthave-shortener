package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
		var html string
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if html, err = h.s.NewShort(ctx, string(url)); err != nil && !errors.Is(err, myErr.ErrAlreadyExist) {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error create new short")
		}
		c.Header("Content-Type", "text/plain; charset=utf-8")
		status := http.StatusCreated
		if errors.Is(err, myErr.ErrAlreadyExist) {
			status = http.StatusConflict
		}
		c.String(status, html)
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
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if result.Result, err = h.s.NewShort(ctx, url.URL); err != nil && !errors.Is(err, myErr.ErrAlreadyExist) {
			c.AbortWithStatus(http.StatusInternalServerError)
			h.log.WithField("Error", err).Error("Error create new short")
		}
		status := http.StatusCreated
		if errors.Is(err, myErr.ErrAlreadyExist) {
			status = http.StatusConflict
		}
		c.JSON(status, result)
	}
}

func (h *Handler) MakeShortBatch() func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			input  []domain.ShortBatchInputItem
			result []domain.ShortBatchResultItem
			err    error
			body   []byte
		)

		if body, err = c.GetRawData(); err != nil || len(body) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err = ffjson.NewDecoder().Decode(body, &input); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if result, err = h.s.NewShortBatch(ctx, input); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				c.String(http.StatusBadRequest, err.Error())
				return
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
				h.log.WithField("Error", err).Error("Error create new batch shorts")
				return
			}
		}
		c.JSON(http.StatusCreated, result)
	}
}

func (h *Handler) GetShort() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if newURL, err := h.s.GetFromShort(ctx, c.Param("id")); err != nil {
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
			h.log.Error("Error ", err)
		} else {
			c.String(http.StatusOK, "Status: ok")
		}
	}
}
