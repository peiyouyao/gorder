package middleware

import (
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HTTPRequestLog(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 打印请求进入日志
		l.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"query":      query,
			"client_ip":  clientIP,
			"user_agent": userAgent,
		}).Info("HTTP request in")

		// 恢复 panic，避免程序崩溃
		defer func() {
			if rec := recover(); rec != nil {
				l.WithFields(logrus.Fields{
					"event":     "panic",
					"method":    method,
					"path":      path,
					"client_ip": clientIP,
					"error":     rec,
					"stack":     string(debug.Stack()),
				}).Error("HTTP request panic")
				// 继续抛出 panic，让 Gin 的 Recover 中间件处理
				panic(rec)
			}
		}()

		// 等下一个中间件 / handler 执行
		c.Next()

		// 请求处理耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 打印请求结束日志
		entry := l.WithFields(logrus.Fields{
			"method":      method,
			"path":        path,
			"query":       query,
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"status_code": statusCode,
			"latency_ms":  latency.Milliseconds(),
		})

		out := "HTTP request out"
		if statusCode >= 500 {
			entry.Error(out + " with server error")
		} else if statusCode >= 400 {
			entry.Warn(out + "with client error")
		} else {
			entry.Info(out)
		}
	}
}
