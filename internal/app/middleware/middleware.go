package middleware

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
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
			_, _ = io.WriteString(c.Writer, err.Error())
			return
		}
		defer func() { _ = gz.Close() }()

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gz}
		c.Next()
	}
}

func Decompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
			if gz, err := gzip.NewReader(c.Request.Body); err == nil {
				c.Request.Body = gz
				defer func() { _ = gz.Close() }()
			}
		}
		c.Next()
	}
}
