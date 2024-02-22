package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func Compress(level int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}
		gz, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			logrus.WithError(err).Error("gzip")
			c.Next()
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				logrus.WithError(err).Error("gzip")
			}
		}()

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gz}
		c.Next()
	}
}

func Decompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(c.Request.Body)
			if err == nil {
				c.Request.Body = gz
				err = gz.Close()
			}
			if err != nil {
				logrus.WithError(err).Error("gzip")
			}
		}
		c.Next()
	}
}
