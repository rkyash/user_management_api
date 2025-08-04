package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details after processing
		duration := time.Since(startTime)
		userID, _ := c.Get("userID")

		entry := logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration,
			"ip":         c.ClientIP(),
			"user-agent": c.Request.UserAgent(),
			"user-id":    userID,
		})

		if c.Writer.Status() >= 500 {
			entry.Error("Server error")
		} else if c.Writer.Status() >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request processed")
		}
	}
}
