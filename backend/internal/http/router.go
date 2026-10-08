package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/AyushCN/harbor/internal/auth"
	"github.com/AyushCN/harbor/internal/config"
	"github.com/AyushCN/harbor/internal/database"
	"github.com/AyushCN/harbor/internal/presence"
	"github.com/AyushCN/harbor/internal/worker"
)

func NewRouter(cfg *config.Config, pool *database.Pool, pm *presence.Manager, nc *nats.Conn) *gin.Engine {
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestLogger())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add config, db, presence manager, and NATS to context for all routes
	r.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Set("db", pool)
		c.Set("presence_manager", pm)
		c.Set("nats", nc)
		c.Next()
	})

	// Health check (no auth)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no auth required)
		v1.POST("/environments/branches", getRepositoryBranches)

		// Public auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/github", handleGitHubLogin)
			authGroup.GET("/github/callback", handleGitHubCallback)
			authGroup.POST("/logout", handleLogout)
			authGroup.GET("/me", auth.AuthRequired(), handleMe)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(auth.AuthRequired())
		{
			// Environments
			envs := protected.Group("/environments")
			{
				envs.GET("", listEnvironments)
				envs.POST("", createEnvironment)
				envs.GET("/:id", getEnvironment)
				envs.POST("/:id/start", startEnvironment)
				envs.POST("/:id/stop", stopEnvironment)
				envs.POST("/:id/resume", resumeEnvironment)
				envs.DELETE("/:id", deleteEnvironment)

				// Members
				envs.GET("/:id/members", listMembers)
				envs.POST("/:id/members", inviteMember)
				envs.PATCH("/:id/members/:user_id", updateMemberRole)
				envs.DELETE("/:id/members/:user_id", removeMember)

				// Share links
				envs.POST("/:id/share-links", createShareLink)

				// Presence
				envs.GET("/:id/presence", getPresence)
				envs.GET("/:id/presence/ws", handlePresenceWS)
			}

			// Share links (public join)
			v1.POST("/share/:token/join", auth.AuthRequired(), joinViaShareLink)
		}
	}

	return r
}

func publishJob(c *gin.Context, jobType worker.JobType, environmentID uuid.UUID, payload interface{}) error {
	nc := c.MustGet("nats").(*nats.Conn)
	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	var payloadBytes []byte
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	job := worker.Job{
		Type:           jobType,
		EnvironmentID:  environmentID,
		Payload:        payloadBytes,
		Timestamp:      time.Now(),
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}

	subject := "env.jobs." + string(jobType)
	_, err = js.Publish(subject, jobBytes)
	return err
}