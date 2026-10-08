package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/AyushCN/harbor/internal/config"
)

type Client struct {
	cli       *client.Client
	cfg       *config.Config
	networkID string
	git       *GitClient
}

type BuildOptions struct {
	WorkspaceID   string
	GitURL        string
	GitBranch     string
	EnvironmentID string
	LogCallback   func(level, message string) // level: info, warn, error, success
}

func NewClient(cfg *config.Config) (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	c := &Client{
		cli: cli,
		cfg: cfg,
		git: NewGitClient(cfg),
	}

	// Ensure network exists
	if err := c.ensureNetwork(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure network: %w", err)
	}

	return c, nil
}

func (c *Client) ensureNetwork(ctx context.Context) error {
	resp, err := c.cli.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, n := range resp.Items {
		if n.Name == c.cfg.DockerNetwork {
			c.networkID = n.ID
			return nil
		}
	}

	// Create network
	ncResp, err := c.cli.NetworkCreate(ctx, c.cfg.DockerNetwork, client.NetworkCreateOptions{
		Driver:  "bridge",
		Labels:  map[string]string{"harbor.managed": "true"},
	})
	if err != nil {
		return err
	}
	c.networkID = ncResp.ID
	log.Printf("Created Docker network: %s", c.networkID)
	return nil
}

func (c *Client) BuildAndRun(ctx context.Context, opts BuildOptions) (string, int, error) {
	emit := func(level, msg string) {
		if opts.LogCallback != nil {
			opts.LogCallback(level, msg)
		}
	}

	// Step 1: Clone repository
	emit("info", fmt.Sprintf("Cloning repository for workspace %s...", opts.WorkspaceID))
	log.Printf("Cloning repository for workspace %s...", opts.WorkspaceID)
	workspacePath, err := c.git.CloneRepository(ctx, opts.WorkspaceID, opts.GitURL, opts.GitBranch)
	if err != nil {
		emit("error", fmt.Sprintf("Clone failed: %v", err))
		return "", 0, fmt.Errorf("clone repository: %w", err)
	}
	emit("success", fmt.Sprintf("Repository cloned successfully"))

	// Step 2: Detect language/framework and generate Dockerfile
	language, dockerfile, err := c.detectLanguageAndDockerfile(workspacePath)
	if err != nil {
		emit("error", fmt.Sprintf("Language detection failed: %v", err))
		return "", 0, fmt.Errorf("detect language: %w", err)
	}
	emit("info", fmt.Sprintf("Detected language: %s", language))

	// Step 3: Create build context from cloned repository
	emit("info", "Creating build context...")
	buildCtx, err := c.git.CreateBuildContext(workspacePath)
	if err != nil {
		emit("error", fmt.Sprintf("Create build context failed: %v", err))
		return "", 0, fmt.Errorf("create build context: %w", err)
	}
	defer buildCtx.Close()

	// Step 4: Build Docker image
	imageName := fmt.Sprintf("harbor-env-%s", opts.EnvironmentID[:12])
	emit("info", fmt.Sprintf("Building image for %s (%s)...", opts.EnvironmentID, language))
	
	// Write Dockerfile to workspace
	dockerfilePath := filepath.Join(workspacePath, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		emit("error", fmt.Sprintf("Write Dockerfile failed: %v", err))
		return "", 0, fmt.Errorf("write Dockerfile: %w", err)
	}

	buildResp, err := c.cli.ImageBuild(ctx, buildCtx, client.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		emit("error", fmt.Sprintf("Image build failed: %v", err))
		return "", 0, fmt.Errorf("image build failed: %w", err)
	}
	defer buildResp.Body.Close()
	
	// Read build output and stream it
	scanner := bufio.NewScanner(buildResp.Body)
	for scanner.Scan() {
		var buildOutput map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &buildOutput); err == nil {
			if stream, ok := buildOutput["stream"].(string); ok {
				emit("info", strings.TrimSpace(stream))
			}
			if errorMsg, ok := buildOutput["error"].(string); ok {
				emit("error", strings.TrimSpace(errorMsg))
			}
		}
	}
	
	emit("success", fmt.Sprintf("Image built successfully: %s", imageName))

	// Step 5: Run container
	containerName := fmt.Sprintf("harbor-%s", opts.EnvironmentID[:12])
	
	// Expose port 3000 (will be detected at runtime)
	exposedPort, _ := network.PortFrom(3000, network.TCP)
	
	resp, err := c.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:        imageName,
			ExposedPorts: map[network.Port]struct{}{
				exposedPort: {},
			},
			Env: []string{
				fmt.Sprintf("HARBOR_ENVIRONMENT_ID=%s", opts.EnvironmentID),
				fmt.Sprintf("HARBOR_WORKSPACE_ID=%s", opts.WorkspaceID),
				"PORT=3000",
			},
			Labels: map[string]string{
				"harbor.environment_id": opts.EnvironmentID,
				"harbor.workspace_id":   opts.WorkspaceID,
				"harbor.managed":        "true",
			},
		},
		HostConfig: &container.HostConfig{
			NetworkMode: container.NetworkMode(c.cfg.DockerNetwork),
			RestartPolicy: container.RestartPolicy{
				Name: "unless-stopped",
			},
			Resources: container.Resources{
				Memory:    512 * 1024 * 1024, // 512MB
				CPUQuota:  50000,             // 0.5 CPU
				PidsLimit: &[]int64{100}[0],
			},
			Tmpfs: map[string]string{
				"/tmp":     "rw,noexec,nosuid,size=100m",
				"/run":     "rw,noexec,nosuid,size=10m",
			},
			SecurityOpt:      []string{"no-new-privileges"},
			ReadonlyRootfs:   true,
			CapDrop:          []string{"ALL"},
		},
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				c.cfg.DockerNetwork: {},
			},
		},
		Name: containerName,
	})
	if err != nil {
		return "", 0, fmt.Errorf("container create: %w", err)
	}

	containerID := resp.ID

	// Start container
	if _, err := c.cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{}); err != nil {
		c.cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true})
		return "", 0, fmt.Errorf("container start: %w", err)
	}

	// Wait for container to be ready
	// Since Traefik reaches containers directly over the Docker network,
	// we just wait for the container to be running, not for host port binding
	if err := c.waitForPort(ctx, containerID, "3000"); err != nil {
		_, _ = c.cli.ContainerStop(ctx, containerID, client.ContainerStopOptions{Timeout: &[]int{10}[0]})
		_, _ = c.cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true})
		return "", 0, fmt.Errorf("wait for container: %w", err)
	}

	return containerID, 0, nil
}

func (c *Client) detectLanguageAndDockerfile(workspacePath string) (string, string, error) {
	// Check for package.json (Node.js)
	packageJSON := filepath.Join(workspacePath, "package.json")
	if _, err := os.Stat(packageJSON); err == nil {
		// Read package.json to check for start script and main entry
		content, _ := os.ReadFile(packageJSON)
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
			Main    string            `json:"main"`
		}
		json.Unmarshal(content, &pkg)
		
		
		// Determine entry point
		entryPoint := "server.js"
		if pkg.Main != "" {
			entryPoint = pkg.Main
		} else {
			// Check for common entry points
			commonEntries := []string{"index.js", "app.js", "main.js", "server.js", "src/index.js", "src/app.js", "src/main.js"}
			for _, entry := range commonEntries {
				if _, err := os.Stat(filepath.Join(workspacePath, entry)); err == nil {
					entryPoint = entry
					break
				}
			}
		}
		
		dockerfile := `# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Runtime stage
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
USER node
EXPOSE 3000
`
		if pkg.Scripts != nil && pkg.Scripts["start"] != "" {
			dockerfile += `CMD ["npm", "start"]`
		} else {
			dockerfile += fmt.Sprintf(`CMD ["node", "%s"]`, entryPoint)
		}
		return "node", dockerfile, nil
	}

	// Check for requirements.txt (Python)
	reqTxt := filepath.Join(workspacePath, "requirements.txt")
	if _, err := os.Stat(reqTxt); err == nil {
		dockerfile := `# Build stage
FROM python:3.11-alpine AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --user -r requirements.txt

# Runtime stage
FROM python:3.11-alpine
WORKDIR /app
COPY --from=builder /root/.local /root/.local
COPY . .
ENV PATH=/root/.local/bin:$PATH
EXPOSE 3000
CMD ["python", "app.py"]
`
		return "python", dockerfile, nil
	}

	// Check for go.mod (Go)
	goMod := filepath.Join(workspacePath, "go.mod")
	if _, err := os.Stat(goMod); err == nil {
		dockerfile := `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Runtime stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 3000
CMD ["./server"]
`
		return "go", dockerfile, nil
	}

	// Check for Cargo.toml (Rust)
	cargoToml := filepath.Join(workspacePath, "Cargo.toml")
	if _, err := os.Stat(cargoToml); err == nil {
		dockerfile := `# Build stage
FROM rust:1.78-alpine AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release 2>/dev/null || true
COPY . .
RUN cargo build --release

# Runtime stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/target/release/server .
EXPOSE 3000
CMD ["./server"]
`
		return "rust", dockerfile, nil
	}

	// Check for pom.xml (Java/Maven)
	pomXML := filepath.Join(workspacePath, "pom.xml")
	if _, err := os.Stat(pomXML); err == nil {
		dockerfile := `# Build stage
FROM maven:3.9-eclipse-temurin-21 AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn clean package -DskipTests

# Runtime stage
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 3000
CMD ["java", "-jar", "app.jar"]
`
		return "java", dockerfile, nil
	}

	// Default to Node.js
	dockerfile := `# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Runtime stage
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
USER node
EXPOSE 3000
CMD ["node", "server.js"]
`
	return "node", dockerfile, nil
}

// Wait for container to be ready
// Since Traefik reaches containers directly over the Docker network,
// we just wait for the container to be running, not for host port binding
func (c *Client) waitForPort(ctx context.Context, containerID string, exposedPort string) error {
	for i := 0; i < 30; i++ {
		inspect, err := c.cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
		if err != nil {
			return err
		}

		if inspect.Container.State.Running {
			// Container is running - Traefik reaches it over the Docker network
			// using the harbor.* labels we set. No host port binding needed.
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for container to start")
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	_, err := c.cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{})
	return err
}

func (c *Client) StopContainer(ctx context.Context, containerID string, timeoutSeconds int) error {
	timeout := timeoutSeconds
	_, err := c.cli.ContainerStop(ctx, containerID, client.ContainerStopOptions{Timeout: &timeout})
	return err
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string) error {
	_, err := c.cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true})
	return err
}

func (c *Client) Close() error {
	return c.cli.Close()
}