package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/AyushCN/harbor/internal/auth"
	"github.com/AyushCN/harbor/internal/database"
	"github.com/AyushCN/harbor/internal/presence"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type PresenceWSMessage struct {
	Type           string                 `json:"type"`
	EnvironmentID  string                 `json:"environment_id,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	Username       string                 `json:"username,omitempty"`
	AvatarURL      string                 `json:"avatar_url,omitempty"`
	Role           string                 `json:"role,omitempty"`
	Timestamp      string                 `json:"timestamp,omitempty"`
	Users          []presence.UserPresence `json:"users,omitempty"`
}

func handlePresenceWS(c *gin.Context) {
	envIDStr := c.Param("id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid environment ID"},
		})
		return
	}

	userID, _ := auth.GetUserID(c)
	username, _ := auth.GetUsername(c)
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

	// Get user info for avatar
	user, _ := queries.GetUserByID(c.Request.Context(), userID)
	avatarURL := ""
	if user != nil && user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Get presence manager from context
	pm := c.MustGet("presence_manager").(*presence.Manager)

	// Subscribe to presence events for this environment
	eventCh := pm.Subscribe(envID.String())
	defer pm.Unsubscribe(envID.String(), eventCh)

	// Join presence
	ctx := c.Request.Context()
	if err := pm.Join(ctx, envID.String(), userID.String(), username, avatarURL, role); err != nil {
		log.Printf("Failed to join presence: %v", err)
	}

	// Send initial sync
	users, _ := pm.GetPresence(ctx, envID.String())
	syncMsg := PresenceWSMessage{
		Type:           "sync",
		EnvironmentID:  envID.String(),
		Users:          users,
		Timestamp:      time.Now().Format(time.RFC3339),
	}
	if err := ws.WriteJSON(syncMsg); err != nil {
		log.Printf("Failed to send sync: %v", err)
		return
	}

	// Handle incoming messages (ping/pong, etc.)
	go func() {
		for {
			var msg PresenceWSMessage
			if err := ws.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}
			// Handle ping/pong
			if msg.Type == "ping" {
				pongMsg := PresenceWSMessage{
					Type:      "pong",
					Timestamp: time.Now().Format(time.RFC3339),
				}
				ws.WriteJSON(pongMsg)
			}
		}
	}()

	// Handle presence events from Redis
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-eventCh:
			msg := PresenceWSMessage{
				Type:          event.Type,
				EnvironmentID: event.EnvironmentID,
				UserID:        event.UserID,
				Username:      event.Username,
				AvatarURL:     event.AvatarURL,
				Role:          event.Role,
				Timestamp:     event.Timestamp.Format(time.RFC3339),
			}
			if err := ws.WriteJSON(msg); err != nil {
				log.Printf("Failed to send presence event: %v", err)
				return
			}
		case <-ticker.C:
			// Send ping to keep connection alive
			pingMsg := PresenceWSMessage{
				Type:      "ping",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			if err := ws.WriteJSON(pingMsg); err != nil {
				return
			}
		case <-c.Request.Context().Done():
			// Leave presence on disconnect
			pm.Leave(c.Request.Context(), envID.String(), userID.String())
			return
		}
	}
}