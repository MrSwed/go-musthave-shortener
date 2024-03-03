package handler

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pquerna/ffjson/ffjson"
)

func (h *Handler) MakeShort() gin.HandlerFunc {
	return func(c *gin.Context) {
		url, err := c.GetRawData()
		if len(url) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithField("Error", err).Error("Error get body")
			return
		}
		var html string
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if html, err = h.s.NewShort(ctx, string(url)); err != nil && !errors.Is(err, myErr.ErrAlreadyExist) {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithField("Error", err).Error("Error create new short")
		}
		c.Header("Content-Type", "text/plain; charset=utf-8")
		status := http.StatusCreated
		if errors.Is(err, myErr.ErrAlreadyExist) {
			status = http.StatusConflict
		}
		c.String(status, html)
	}
}

func (h *Handler) MakeShortJSON() gin.HandlerFunc {
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
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if result.Result, err = h.s.NewShort(ctx, url.URL); err != nil && !errors.Is(err, myErr.ErrAlreadyExist) {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithField("Error", err).Error("Error create new short")
		}
		status := http.StatusCreated
		if errors.Is(err, myErr.ErrAlreadyExist) {
			status = http.StatusConflict
		}
		c.JSON(status, result)
	}
}

func (h *Handler) MakeShortBatch() gin.HandlerFunc {
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
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if result, err = h.s.NewShortBatch(ctx, input); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				c.String(http.StatusBadRequest, err.Error())
				return
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
				logrus.WithField("Error", err).Error("Error create new batch shorts")
				return
			}
		}
		c.JSON(http.StatusCreated, result)
	}
}

func (h *Handler) GetShort() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if newURL, err := h.s.GetFromShort(ctx, c.Param("id")); err != nil {
			switch true {
			case errors.Is(err, myErr.ErrIsDeleted):
				c.AbortWithStatus(http.StatusGone)
			case errors.Is(err, myErr.ErrNotExist):
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				c.AbortWithStatus(http.StatusInternalServerError)
				logrus.WithField("Error", err).Error("Error get new short")
			}
			return
		} else {
			c.Redirect(http.StatusTemporaryRedirect, newURL)
			return
		}
	}
}

func (h *Handler) GetDBPing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		if err := h.s.CheckDB(ctx); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.Error("Error ", err)
		} else {
			c.String(http.StatusOK, "Status: ok")
		}
	}
}

func (h *Handler) GetAllByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		userID := ""
		if u, ok := ctx.Value(constant.ContextUserValueName).(string); ok {
			userID = u
		}
		if userID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		data, err := h.s.GetAllByUser(ctx, userID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.Error("Error ", err)
			return
		}
		status := http.StatusOK
		if len(data) == 0 {
			status = http.StatusNoContent
		}
		c.JSON(status, data)
	}
}

func (h *Handler) SetDeleted() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		userID := ""
		if u, ok := ctx.Value(constant.ContextUserValueName).(string); ok {
			userID = u
		}
		if userID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var (
			inputShorts []string
			err         error
			body        []byte
		)
		if body, err = c.GetRawData(); err != nil || len(body) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err = ffjson.NewDecoder().Decode(body, &inputShorts); err != nil || len(inputShorts) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		_, err = h.s.SetDeleted(ctx, userID, true, inputShorts...)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"userID": userID, "delete": true, "shorts": inputShorts}).Error(err)
			return
		}
		c.Status(http.StatusAccepted)
	}
}
