package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/harbor/internal/auth"
	"github.com/yourusername/harbor/internal/database"
)

func handleGitHubLogin(c *gin.Context) {
	auth.HandleGitHubLogin(c)
}

func handleGitHubCallback(c *gin.Context) {
	auth.HandleGitHubCallback(c)
}

func handleLogout(c *gin.Context) {
	// Clear the auth cookie with matching attributes
	c.SetCookie("harbor_token", "", -1, "/", "", false, true)
	// Also clear any refresh token cookie if present
	c.SetCookie("harbor_refresh", "", -1, "/", "", false, true)
	
	// TODO: Add token blacklist/revocation in database if needed
	// For now, client-side token invalidation is handled by frontend
	
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func handleMe(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	username, _ := auth.GetUsername(c)

	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	user, err := queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"id":       userID,
			"username": username,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"name":        user.Name,
		"avatar_url":  user.AvatarURL,
	})
}