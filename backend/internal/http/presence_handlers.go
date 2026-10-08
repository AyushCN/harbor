package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/AyushCN/harbor/internal/auth"
	"github.com/AyushCN/harbor/internal/database"
)

type PresenceUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  string    `json:"joined_at"`
}

func getPresence(c *gin.Context) {
	envIDStr := c.Param("id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid environment ID"},
		})
		return
	}

	userID, _ := auth.GetUserID(c)
	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	// Check access
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "You don't have access to this environment"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to check access"},
		})
		return
	}

	// TODO: Query Redis for current presence
	// For now return empty list with current user
	c.JSON(http.StatusOK, gin.H{
		"environment_id": envID,
		"current_user": gin.H{
			"id":       userID,
			"role":     role,
		},
		"users": []PresenceUser{},
		"server_time": time.Now().Format(time.RFC3339),
	})
}