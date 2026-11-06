package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

const (
	RequestIDHeader     = "X-Request-ID"
	ContextKeyRequestID = "request_id"
)

func RequestID() echo.MiddlewareFunc {
	return echoMiddleware.RequestIDWithConfig(echoMiddleware.RequestIDConfig{
		TargetHeader: RequestIDHeader,
		Generator:    uuid.NewString,
		RequestIDHandler: func(c *echo.Context, requestID string) {
			c.Request().Header.Set(RequestIDHeader, requestID)
			c.Set(ContextKeyRequestID, requestID)
		},
	})
}

func RequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogContentLength: true,
		LogResponseSize:  true,
		HandleError:      true,
		LogValuesFunc: func(c *echo.Context, v echoMiddleware.RequestLoggerValues) error {
			fields := []zap.Field{
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.String("host", v.Host),
				zap.String("bytes_in", v.ContentLength),
				zap.Int64("bytes_out", v.ResponseSize),
				zap.String("user_agent", v.UserAgent),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("request_id", v.RequestID),
			}

			if v.Error != nil {
				logger.Error("request error", append(fields, zap.Error(v.Error))...)
				return nil
			}

			logger.Info("request", fields...)
			return nil
		},
	})
}
