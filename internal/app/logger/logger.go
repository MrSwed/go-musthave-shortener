package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger(l logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
			l.WithFields(logrus.Fields{
				"status":   c.Writer.Status(),
				"method":   c.Request.Method,
				"URI":      fmt.Sprintf("%s://%s%s %s", scheme, c.Request.Host, c.Request.RequestURI, c.Request.Proto),
				"size":     c.Writer.Size(),
				"duration": time.Since(start),
				"from":     c.Request.RemoteAddr,
			}).Info("Served")
		}()
		c.Next()
	}
}
