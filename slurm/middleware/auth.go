package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查端点
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// 获取认证信息（这里简化处理，实际应用中应该有更完善的认证机制）
		sessionKey := c.GetHeader("sessionKey")
		authorization := c.GetHeader("Authorization")

		// 检查 sessionKey 或 Authorization
		if sessionKey == "" && authorization == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please provide sessionKey header or Authorization header",
			})
			c.Abort()
			return
		}

		// 简单的 token 验证（实际应用中应该验证 token 的有效性）
		if authorization != "" {
			if !strings.HasPrefix(authorization, "Bearer ") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Invalid authorization format",
					"message": "Authorization header must use Bearer token format",
				})
				c.Abort()
				return
			}
		}

		// 将认证信息存储到上下文中
		if sessionKey != "" {
			c.Set("sessionKey", sessionKey)
		}

		if authorization != "" {
			token := strings.TrimPrefix(authorization, "Bearer ")
			c.Set("token", token)
		}

		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Requested-With, sessionKey")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
