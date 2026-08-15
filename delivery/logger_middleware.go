package delivery

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-ID"

// InitLogger menginisialisasi structured JSON logger bawaan Go 1.21+
func InitLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// RequestIDMiddleware menyematkan atau membuat UUID unik X-Request-ID untuk setiap request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Simpan di context Gin dan sertakan di Header Response
		c.Set(HeaderXRequestID, reqID)
		c.Header(HeaderXRequestID, reqID)

		c.Next()
	}
}

// StructuredLoggerMiddleware mencatat log HTTP request dalam format JSON
func StructuredLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		reqID, _ := c.Get(HeaderXRequestID)

		if raw != "" {
			path = path + "?" + raw
		}

		// Tentukan level log berdasarkan status code
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		logger.Log(c.Request.Context(), logLevel, "HTTP Request Log",
			slog.String("request_id", reqID.(string)),
			slog.Int("status", statusCode),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
