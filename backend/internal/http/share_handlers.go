package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/AyushCN/harbor/internal/auth"
	"github.com/AyushCN/harbor/internal/config"
	"github.com/AyushCN/harbor/internal/database"
)

type CreateShareLinkRequest struct {
	Role          string `json:"role" binding:"required,oneof=COLLABORATOR VIEWER"`
	ExpiresInHours int   `json:"expires_in_hours"`
	MaxUses       int    `json:"max_uses"`
}

func createShareLink(c *gin.Context) {
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

	// Check if user is Owner
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil || role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can create share links"},
		})
		return
	}

	var req CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	// Generate secure token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate token"},
		})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	var expiresAt *time.Time
	if req.ExpiresInHours > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	var maxUses *int
	if req.MaxUses > 0 {
		maxUses = &req.MaxUses
	}

	shareLink, err := queries.CreateShareLink(c.Request.Context(), envID, userID, token, req.Role, expiresAt, maxUses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create share link"},
		})
		return
	}

	cfg := c.MustGet("config").(*config.Config)

	// Use config frontend URL
	shareURL := cfg.FrontendURL + "/share/" + token

	c.JSON(http.StatusCreated, gin.H{
		"token":      shareLink.Token,
		"url":        shareURL,
		"role":       shareLink.Role,
		"expires_at": shareLink.ExpiresAt,
		"max_uses":   shareLink.MaxUses,
	})
}

func joinViaShareLink(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Missing share token"},
		})
		return
	}

	userID, _ := auth.GetUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"},
		})
		return
	}

	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	shareLink, err := queries.GetShareLinkByToken(c.Request.Context(), token)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "Invalid or expired share link"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to validate share link"},
		})
		return
	}

	// Check expiry
	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{
			"error": gin.H{"code": "EXPIRED", "message": "Share link has expired"},
		})
		return
	}

	// Check max uses
	if shareLink.MaxUses != nil && shareLink.UseCount >= *shareLink.MaxUses {
		c.JSON(http.StatusGone, gin.H{
			"error": gin.H{"code": "EXPIRED", "message": "Share link has reached maximum uses"},
		})
		return
	}

	// Add user as member
	_, err = queries.AddEnvironmentMember(c.Request.Context(), shareLink.EnvironmentID, userID, shareLink.Role, &shareLink.CreatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to join environment"},
		})
		return
	}

	// Increment use count
	_ = queries.IncrementShareLinkUseCount(c.Request.Context(), token)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined environment",
		"environment_id": shareLink.EnvironmentID,
	})
}