package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/harbor/internal/config"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		// Try to get token from header
		authHeader := c.GetHeader("Authorization")
		tokenString := ExtractTokenFromHeader(authHeader)

		// If no header, try cookie
		if tokenString == "" {
			cookie, err := c.Cookie("harbor_token")
			if err == nil {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Missing authentication token"},
			})
			return
		}

		claims, err := ValidateToken(tokenString, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid or expired token"},
			})
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func GetUsername(c *gin.Context) (string, bool) {
	val, exists := c.Get("username")
	if !exists {
		return "", false
	}
	username, ok := val.(string)
	return username, ok
}