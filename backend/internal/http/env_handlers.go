package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/harbor/internal/auth"
	"github.com/yourusername/harbor/internal/config"
	"github.com/yourusername/harbor/internal/database"
	"github.com/yourusername/harbor/internal/worker"
)

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type CreateEnvironmentRequest struct {
	GitURL   string `json:"git_url" binding:"required"`
	GitBranch string `json:"git_branch"`
	Name     string `json:"name"`
}

type EnvironmentResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	PublicURL string    `json:"public_url,omitempty"`
	Host      *UserInfo `json:"host"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"created_at"`
	GitURL    string    `json:"git_url,omitempty"`
	GitBranch string    `json:"git_branch,omitempty"`
}

func listEnvironments(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	envs, err := queries.GetUserEnvironmentsWithHost(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list environments"},
		})
		return
	}

	var resp []EnvironmentResponse
	for _, e := range envs {
		name := e.WorkspaceName
		if name == "" {
			name = e.WorkspaceGitURL
		}
		resp = append(resp, EnvironmentResponse{
			ID:        e.ID,
			Name:      name,
			Status:    e.Status,
			Host: &UserInfo{
				ID:        e.Host.ID,
				Username:  e.Host.Username,
				AvatarURL: e.Host.AvatarURL,
			},
			Role:      e.Role,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
			GitURL:    e.WorkspaceGitURL,
			GitBranch: e.WorkspaceGitBranch,
			PublicURL: ptrToString(e.PublicURL),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func createEnvironment(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	gitBranch := req.GitBranch
	if gitBranch == "" {
		gitBranch = "main"
	}

	// Create workspace
	workspace, err := queries.CreateWorkspace(c.Request.Context(), userID, req.GitURL, gitBranch, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create workspace: " + err.Error()},
		})
		return
	}

	// Create environment
	env, err := queries.CreateEnvironment(c.Request.Context(), workspace.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create environment: " + err.Error()},
		})
		return
	}

	// Add owner as member
	_, err = queries.AddEnvironmentMember(c.Request.Context(), env.ID, userID, "OWNER", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to add owner as member: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      env.ID,
		"status":  env.Status,
		"host_id": userID,
	})

	// Publish create job to NATS
	payload := map[string]string{
		"git_url":    req.GitURL,
		"git_branch": gitBranch,
		"name":       req.Name,
	}
	if err := publishJob(c, worker.JobCreate, env.ID, payload); err != nil {
		log.Printf("Failed to publish create job: %v", err)
	}
}

func getEnvironment(c *gin.Context) {
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

	env, err := queries.GetEnvironmentByID(c.Request.Context(), envID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "Environment not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to fetch environment"},
		})
		return
	}

	// Get workspace for name
	workspace, _ := queries.GetWorkspaceByID(c.Request.Context(), env.WorkspaceID)
	name := ""
	if workspace != nil && workspace.Name != nil {
		name = *workspace.Name
	}

	// Get host info
	host, _ := queries.GetUserByID(c.Request.Context(), env.HostID)

	c.JSON(http.StatusOK, gin.H{
		"id":          env.ID,
		"name":        name,
		"status":      env.Status,
		"public_url":  env.PublicURL,
		"host":        host,
		"role":        role,
		"created_at":  env.CreatedAt.Format(time.RFC3339),
		"workspace":   workspace,
	})
}

func startEnvironment(c *gin.Context) {
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

	// Check role (Owner or Collaborator)
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil || (role != "OWNER" && role != "COLLABORATOR") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner or Collaborator can start environment"},
		})
		return
	}

	// Publish start job to NATS
	if err := publishJob(c, worker.JobStart, envID, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to queue start job"},
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Environment start queued"})
}

func stopEnvironment(c *gin.Context) {
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

	// Check role (Owner only)
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil || role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can stop environment"},
		})
		return
	}

	// Publish stop job to NATS
	if err := publishJob(c, worker.JobStop, envID, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to queue stop job"},
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Environment stop queued"})
}

func resumeEnvironment(c *gin.Context) {
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

	// Check role (Owner or Collaborator)
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil || (role != "OWNER" && role != "COLLABORATOR") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner or Collaborator can resume environment"},
		})
		return
	}

	// Publish resume job to NATS
	if err := publishJob(c, worker.JobResume, envID, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to queue resume job"},
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Environment resume queued"})
}

func deleteEnvironment(c *gin.Context) {
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

	// Check role (Owner only)
	role, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
	if err != nil || role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Only Owner can delete environment"},
		})
		return
	}

	// Publish delete job to NATS (worker will handle soft delete after removing container)
	if err := publishJob(c, worker.JobDelete, envID, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to queue delete job"},
		})
		return
	}

	// Update status immediately to indicate deletion is in progress
	queries.UpdateEnvironmentStatus(c.Request.Context(), envID, "DELETING")

	c.JSON(http.StatusAccepted, gin.H{"message": "Environment deletion queued"})
}
// BranchResponse represents a GitHub branch
type BranchResponse struct {
	Name   string `json:"name"`
	SHA    string `json:"sha"`
	Protected bool `json:"protected"`
}

// GetBranchesResponse for API response
type GetBranchesResponse struct {
	Branches []BranchResponse `json:"branches"`
	DefaultBranch string `json:"default_branch"`
}

// getRepositoryBranches fetches branches from a GitHub repository
func getRepositoryBranches(c *gin.Context) {
	var req struct {
		GitURL string `json:"git_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Git URL is required"},
		})
		return
	}

	// Parse GitHub URL to extract owner and repo
	owner, repo, err := parseGitHubURL(req.GitURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid GitHub URL: " + err.Error()},
		})
		return
	}

	// Try to get branches using GitHub API (no auth first, then with token if available)
	branches, defaultBranch, err := fetchGitHubBranches(owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "GITHUB_API_ERROR", "message": "Failed to fetch branches: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, GetBranchesResponse{
		Branches:      branches,
		DefaultBranch: defaultBranch,
	})
}

// parseGitHubURL extracts owner and repo from various GitHub URL formats
func parseGitHubURL(url string) (string, string, error) {
	// Handle various formats:
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	// https://github.com/owner/repo/

	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	
	// Remove git@github.com: format
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
	}
	
	// Remove https://github.com/ or http://github.com/
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")
	
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format")
	}
	
	owner := parts[0]
	repo := parts[1]
	
	// Remove any remaining .git suffix
	repo = strings.TrimSuffix(repo, ".git")
	
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid owner or repo name")
	}
	
	return owner, repo, nil
}

// fetchGitHubBranches fetches branches from GitHub API
func fetchGitHubBranches(owner, repo string) ([]BranchResponse, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?per_page=100", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Harbor-Dev-Platform")
	
	// Add GitHub token if available from config (for higher rate limits)
	cfg, _ := config.Load()
	if cfg.GitHubClientID != "" && cfg.GitHubSecret != "" {
		// Could use app installation token here, but for now use public API
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, "", fmt.Errorf("repository not found")
	}
	if resp.StatusCode == 403 {
		return nil, "", fmt.Errorf("rate limited or access denied")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GitHub API error: %s", string(body))
	}
	
	var branches []BranchResponse
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, "", err
	}
	
	// Find default branch (usually the one with "protected": true or first branch)
	defaultBranch := "main"
	for _, branch := range branches {
		if branch.Protected || branch.Name == "main" || branch.Name == "master" {
			defaultBranch = branch.Name
			break
		}
	}
	
	return branches, defaultBranch, nil
}

// ============================================
// BUILD LOGS HANDLERS
// ============================================

type BuildLogResponse struct {
	ID        uuid.UUID `json:"id"`
	Sequence  int       `json:"sequence"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	CreatedAt string    `json:"created_at"`
}

type BuildLogsResponse struct {
	Logs  []BuildLogResponse `json:"logs"`
	Total int                `json:"total"`
}

func getBuildLogs(c *gin.Context) {
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

	// Parse query params
	limit := 1000
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	logs, err := queries.GetBuildLogs(c.Request.Context(), envID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to fetch build logs"},
		})
		return
	}

	total, _ := queries.GetBuildLogCount(c.Request.Context(), envID)

	resp := BuildLogsResponse{
		Logs:  make([]BuildLogResponse, len(logs)),
		Total: total,
	}
	for i, log := range logs {
		resp.Logs[i] = BuildLogResponse{
			ID:        log.ID,
			Sequence:  log.Sequence,
			Message:   log.Message,
			Level:     log.Level,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, resp)
}

func handleBuildLogsWS(c *gin.Context) {
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
	_, err := queries.GetUserRoleInEnvironment(c.Request.Context(), envID, userID)
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

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Send historical logs first
	logs, _ := queries.GetBuildLogs(c.Request.Context(), envID, 500, 0)
	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]
		msg := map[string]any{
			"type":       "log",
			"id":         log.ID,
			"sequence":   log.Sequence,
			"message":    log.Message,
			"level":      log.Level,
			"created_at": log.CreatedAt.Format(time.RFC3339),
			"historical": true,
		}
		if err := ws.WriteJSON(msg); err != nil {
			return
		}
	}

	// Subscribe to NATS for real-time logs
	nc := c.MustGet("nats").(*nats.Conn)
	js, _ := nc.JetStream()

	sub, err := js.PullSubscribe("build.logs."+envIDStr, "ws-"+uuid.New().String(), nats.BindStream("BUILD_LOGS"))
	if err != nil {
		log.Printf("Failed to subscribe to build logs: %v", err)
		return
	}
	defer sub.Unsubscribe()

	// Send logs from NATS to WebSocket
	for {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(500*time.Millisecond))
			if err != nil {
				if err != nats.ErrTimeout {
					log.Printf("Fetch error for build logs: %v", err)
				}
				continue
			}

			for _, msg := range msgs {
				var logMsg map[string]any
				if err := json.Unmarshal(msg.Data, &logMsg); err != nil {
					msg.Nak()
					continue
				}
				logMsg["historical"] = false
				if err := ws.WriteJSON(logMsg); err != nil {
					return
				}
				msg.Ack()
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}
