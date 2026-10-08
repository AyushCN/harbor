package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/harbor/internal/auth"
	"github.com/yourusername/harbor/internal/database"
)

type InviteMemberRequest struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=COLLABORATOR VIEWER"`
}

type MemberResponse struct {
	User      UserInfo `json:"user"`
	Role      string   `json:"role"`
	JoinedAt  string   `json:"joined_at"`
}

func listMembers(c *gin.Context) {
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
	_, err = queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
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

	members, err := queries.GetEnvironmentMembers(c.Request.Context(), envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to fetch members"},
		})
		return
	}

	var resp []MemberResponse
	for _, m := range members {
		resp = append(resp, MemberResponse{
			User: UserInfo{
				ID:        m.User.ID,
				Username:  m.User.Username,
				AvatarURL: m.User.AvatarURL,
			},
			Role:     m.Role,
			JoinedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func inviteMember(c *gin.Context) {
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
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can invite members"},
		})
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	// Find user by GitHub username
	targetUser, err := queries.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "User not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to find user"},
		})
		return
	}

	// Check if already a member
	existingRole, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, targetUser.ID)
	if err == nil && existingRole != "" {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "CONFLICT", "message": "User is already a member"},
		})
		return
	}

	// Add member
	_, err = queries.AddEnvironmentMember(c.Request.Context(), envID, targetUser.ID, req.Role, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to add member"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invitation sent",
		"member": MemberResponse{
			User: UserInfo{
				ID:        targetUser.ID,
				Username:  targetUser.Username,
				AvatarURL: targetUser.AvatarURL,
			},
			Role:     req.Role,
			JoinedAt: time.Now().Format(time.RFC3339),
		},
	})
}

func updateMemberRole(c *gin.Context) {
	envIDStr := c.Param("id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid environment ID"},
		})
		return
	}

	memberIDStr := c.Param("user_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"},
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
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can change roles"},
		})
		return
	}

	// Prevent changing own role
	if memberID == userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Cannot change your own role"},
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=COLLABORATOR VIEWER"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	err = queries.UpdateMemberRole(c.Request.Context(), envID, memberID, req.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "Member not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update role"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated",
		"environment_id": envID,
		"user_id": memberID,
		"role": req.Role,
	})
}

func removeMember(c *gin.Context) {
	envIDStr := c.Param("id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid environment ID"},
		})
		return
	}

	memberIDStr := c.Param("user_id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid user ID"},
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
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can remove members"},
		})
		return
	}

	// Prevent removing self
	if memberID == userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Cannot remove yourself"},
		})
		return
	}

	err = queries.RemoveMember(c.Request.Context(), envID, memberID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "Member not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to remove member"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed",
		"environment_id": envID,
		"user_id": memberID,
	})
}