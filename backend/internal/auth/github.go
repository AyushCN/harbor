package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/harbor/internal/config"
	"github.com/yourusername/harbor/internal/database"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func HandleGitHubLogin(c *gin.Context) {
	cfg := c.MustGet("config").(*config.Config)

	// Generate state for CSRF protection
	state := uuid.New().String()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	// Redirect to GitHub OAuth
	githubURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user,user:email&state=%s",
		cfg.GitHubClientID,
		url.QueryEscape(cfg.APIURL+"/api/v1/auth/github/callback"),
		state,
	)

	c.Redirect(http.StatusTemporaryRedirect, githubURL)
}

func HandleGitHubCallback(c *gin.Context) {
	cfg := c.MustGet("config").(*config.Config)
	pool := c.MustGet("db").(*database.Pool)
	queries := database.NewQueries(pool)

	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=invalid_state")
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=no_code")
		return
	}

	// Exchange code for access token
	tokenResp, err := exchangeCodeForToken(cfg, code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=token_exchange_failed")
		return
	}

	// Get user info from GitHub
	githubUser, err := getGitHubUser(tokenResp.AccessToken)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=user_fetch_failed")
		return
	}

	// Encrypt the access token using AES-256-GCM
	encryptedToken, err := encryptToken(tokenResp.AccessToken, cfg.EncryptionKey)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=encryption_failed")
		return
	}

	// Find or create user
	user, err := queries.GetUserByGithubID(c.Request.Context(), githubUser.ID)
	if err != nil {
		// Create new user
		user, err = queries.CreateUser(c.Request.Context(), githubUser.ID, githubUser.Login, githubUser.Name, githubUser.AvatarURL, encryptedToken)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=user_create_failed")
			return
		}
	} else {
		// Update existing user
		user, err = queries.GetUserByID(c.Request.Context(), user.ID)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=user_fetch_failed")
			return
		}
		// Update token
		_ = queries.UpdateUserToken(c.Request.Context(), user.ID, encryptedToken)
	}

	// Generate JWT
	jwtToken, err := GenerateToken(user.ID, user.Username, cfg)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/?error=token_gen_failed")
		return
	}

	// Set cookie and redirect
	// Use empty domain for localhost (works for both localhost and 127.0.0.1)
	c.SetCookie("harbor_token", jwtToken, 86400, "/", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/dashboard")
}

func exchangeCodeForToken(cfg *config.Config, code string) (*GitHubTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", cfg.GitHubClientID)
	data.Set("client_secret", cfg.GitHubSecret)
	data.Set("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp GitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &tokenResp, nil
}

func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// encryptToken encrypts a token using AES-256-GCM
// key must be 32 bytes (256 bits) - derived from the EncryptionKey config
func encryptToken(plaintext, key string) (string, error) {
	// Derive 32-byte key from the provided key
	keyBytes := deriveKey(key)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken decrypts a token using AES-256-GCM
func decryptToken(ciphertextB64, key string) (string, error) {
	keyBytes := deriveKey(key)

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// deriveKey derives a 32-byte key from the provided string using SHA-256
func deriveKey(key string) []byte {
	// Use the first 32 bytes of SHA-256(key)
	// In production, consider using PBKDF2 or similar
	sum := sha256Sum(key)
	return sum[:32]
}

// sha256Sum returns SHA-256 hash of input
func sha256Sum(input string) []byte {
	h := sha256.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

// Need to add imports for crypto/sha256