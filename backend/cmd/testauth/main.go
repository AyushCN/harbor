package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString := authHeader[7:]
			claims := &Claims{}
			token, _ := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("super-secret-change-me-in-production-at-least-32-chars"), nil
			})
			if token != nil && token.Valid {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
			}
		}
		c.Next()
	})

	r.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "no user_id"})
			return
		}
		id, ok := val.(uuid.UUID)
		if !ok {
			c.JSON(401, gin.H{"error": "invalid type"})
			return
		}
		c.JSON(200, gin.H{"user_id": id})
	})

	r.Run(":8085")
}