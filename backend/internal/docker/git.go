package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AyushCN/harbor/internal/config"
)

type GitClient struct {
	cfg *config.Config
}

func NewGitClient(cfg *config.Config) *GitClient {
	return &GitClient{cfg: cfg}
}

func (g *GitClient) CloneRepository(ctx context.Context, workspaceID, gitURL, gitBranch string) (string, error) {
	workspacePath := filepath.Join(g.cfg.WorkspaceRoot, workspaceID)
	
	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return "", fmt.Errorf("create workspace dir: %w", err)
	}

	// Check if already cloned
	gitDir := filepath.Join(workspacePath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Already cloned, just pull latest
		log.Printf("Repository already exists, pulling latest for %s", workspaceID)
		if err := g.pullLatest(ctx, workspacePath, gitBranch); err != nil {
			return "", fmt.Errorf("pull latest: %w", err)
		}
		return workspacePath, nil
	}

	// Clone repository
	log.Printf("Cloning %s (branch: %s) to %s", gitURL, gitBranch, workspacePath)
	
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--branch", gitBranch, "--depth", "1", gitURL, workspacePath)
	cloneCmd.Dir = workspacePath
	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	log.Printf("Successfully cloned repository for %s", workspaceID)
	return workspacePath, nil
}

func (g *GitClient) pullLatest(ctx context.Context, workspacePath, gitBranch string) error {
	// Fetch and reset to origin/branch
	cmds := [][]string{
		{"git", "fetch", "origin", gitBranch},
		{"git", "reset", "--hard", "origin/" + gitBranch},
		{"git", "clean", "-fd"},
	}

	for _, args := range cmds {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = workspacePath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git %s failed: %w, output: %s", args[0], err, string(output))
		}
	}
	return nil
}

func (g *GitClient) GetCommitHash(ctx context.Context, workspacePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = workspacePath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitClient) CleanupWorkspace(workspaceID string) error {
	workspacePath := filepath.Join(g.cfg.WorkspaceRoot, workspaceID)
	return os.RemoveAll(workspacePath)
}

// DockerBuildContext creates a proper build context from a Git repository
func (g *GitClient) CreateBuildContext(workspacePath string) (io.ReadCloser, error) {
	// Create a tar archive of the workspace for Docker build context
	// This is more efficient than copying files
	
	pr, pw := io.Pipe()
	
	go func() {
		defer pw.Close()
		cmd := exec.Command("tar", "-czf", "-", "-C", workspacePath, ".")
		cmd.Stdout = pw
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("tar command failed: %v", err)
		}
	}()
	
	return pr, nil
}