package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/AyushCN/harbor/internal/config"
	"github.com/AyushCN/harbor/internal/database"
	"github.com/AyushCN/harbor/internal/docker"
)

type JobType string

const (
	JobCreate    JobType = "environment.create"
	JobStart     JobType = "environment.start"
	JobStop      JobType = "environment.stop"
	JobResume    JobType = "environment.resume"
	JobDelete    JobType = "environment.delete"
)

type Job struct {
	Type        JobType      `json:"type"`
	EnvironmentID uuid.UUID  `json:"environment_id"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
}

type CreatePayload struct {
	GitURL   string `json:"git_url"`
	GitBranch string `json:"git_branch"`
	Name     string `json:"name"`
}

type Worker struct {
	cfg        *config.Config
	nc         *nats.Conn
	js         nats.JetStreamContext
	db         *database.Pool
	docker     *docker.Client
	queries    *database.Queries
	shutdownCh chan struct{}
}

func NewWorker(cfg *config.Config) (*Worker, error) {
	// Connect to NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("create JetStream context: %w", err)
	}

	// Initialize database
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Initialize Docker client
	dockerClient, err := docker.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create Docker client: %w", err)
	}

	w := &Worker{
		cfg:        cfg,
		nc:         nc,
		js:         js,
		db:         pool,
		docker:     dockerClient,
		queries:    database.NewQueries(pool),
		shutdownCh: make(chan struct{}),
	}

	return w, nil
}

func (w *Worker) Start(ctx context.Context) error {
	// Create stream for environment jobs
	_, err := w.js.AddStream(&nats.StreamConfig{
		Name:     "ENV_JOBS",
		Subjects: []string{"env.jobs.>"},
		Storage:  nats.FileStorage,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return fmt.Errorf("create stream: %w", err)
	}

	// Create consumer for each job type
	jobTypes := []JobType{JobCreate, JobStart, JobStop, JobResume, JobDelete}
	for _, jobType := range jobTypes {
		durableName := strings.ReplaceAll(string(jobType), ".", "_")
		_, err := w.js.AddConsumer("ENV_JOBS", &nats.ConsumerConfig{
			Durable:       durableName,
			FilterSubject: "env.jobs." + string(jobType),
			AckPolicy:     nats.AckExplicitPolicy,
			MaxDeliver:    3,
			AckWait:       30 * time.Second,
		})
		if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
			return fmt.Errorf("create consumer for %s: %w", jobType, err)
		}
	}

	// Subscribe to each job type
	for _, jobType := range jobTypes {
		go w.consumeJob(jobType)
	}

	log.Println("Worker started, consuming jobs...")
	<-w.shutdownCh
	return nil
}

func (w *Worker) consumeJob(jobType JobType) {
	durableName := strings.ReplaceAll(string(jobType), ".", "_")
	sub, err := w.js.PullSubscribe("env.jobs."+string(jobType), durableName, nats.BindStream("ENV_JOBS"))
	if err != nil {
		log.Printf("Failed to subscribe to %s: %v", jobType, err)
		return
	}

	for {
		select {
		case <-w.shutdownCh:
			sub.Unsubscribe()
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if err != nil {
				if err != nats.ErrTimeout {
					log.Printf("Fetch error for %s: %v", jobType, err)
				}
				continue
			}

			for _, msg := range msgs {
				var job Job
				if err := json.Unmarshal(msg.Data, &job); err != nil {
					log.Printf("Failed to unmarshal job: %v", err)
					msg.Nak()
					continue
				}

				ctx := context.Background()
				var jobErr error

				switch job.Type {
				case JobCreate:
					jobErr = w.handleCreate(ctx, job)
				case JobStart:
					jobErr = w.handleStart(ctx, job)
				case JobStop:
					jobErr = w.handleStop(ctx, job)
				case JobResume:
					jobErr = w.handleResume(ctx, job)
				case JobDelete:
					jobErr = w.handleDelete(ctx, job)
				}

				if jobErr != nil {
					log.Printf("Job %s failed: %v", job.Type, jobErr)
					msg.Nak()
				} else {
					msg.Ack()
				}
			}
		}
	}
}

func (w *Worker) handleCreate(ctx context.Context, job Job) error {
	// Update status to BUILDING
	if err := w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "BUILDING"); err != nil {
		return fmt.Errorf("update status to BUILDING: %w", err)
	}

	// Get environment details
	env, err := w.queries.GetEnvironmentByID(ctx, job.EnvironmentID)
	if err != nil {
		return fmt.Errorf("get environment: %w", err)
	}

	// Get workspace
	workspace, err := w.queries.GetWorkspaceByID(ctx, env.WorkspaceID)
	if err != nil {
		return fmt.Errorf("get workspace: %w", err)
	}

	// Parse payload
	var payload CreatePayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Emit build log
	emitLog := func(level, message string) {
		_ = w.queries.AddBuildLog(ctx, job.EnvironmentID, nil, message, database.BuildLogLevel(level))
		// Also publish to NATS for real-time WebSocket
		logMsg := map[string]any{
			"type":       "log",
			"environment_id": job.EnvironmentID.String(),
			"message":    message,
			"level":      level,
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(logMsg)
		w.nc.Publish("build.logs."+job.EnvironmentID.String(), data)
	}

	// Emit initial log
	emitLog("info", fmt.Sprintf("Starting environment creation for %s", job.EnvironmentID))

	// Build and run container
	containerID, port, err := w.docker.BuildAndRun(ctx, docker.BuildOptions{
		WorkspaceID:   workspace.ID.String(),
		GitURL:        workspace.GitURL,
		GitBranch:     workspace.GitBranch,
		EnvironmentID: job.EnvironmentID.String(),
		LogCallback:   emitLog,
	})
	if err != nil {
		w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "BUILD_FAILED")
		emitLog("error", "✗ BUILD FAILED: "+err.Error())
		return fmt.Errorf("build and run: %w", err)
	}

	publicURL := fmt.Sprintf("http://%s.%s", job.EnvironmentID.String()[:8], w.cfg.TraefikDomain)

	if err := w.queries.UpdateEnvironmentContainer(ctx, job.EnvironmentID, containerID, port, publicURL); err != nil {
		w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "BUILD_FAILED")
		emitLog("error", "✗ Failed to save container info")
		return fmt.Errorf("update environment container: %w", err)
	}

	// Explicit success
	if err := w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "RUNNING"); err != nil {
		return fmt.Errorf("update status to RUNNING: %w", err)
	}

	emitLog("success", "✓ Environment is READY")
	log.Printf("Environment %s created successfully", job.EnvironmentID)
	return nil
}

func (w *Worker) handleStart(ctx context.Context, job Job) error {
	env, err := w.queries.GetEnvironmentByID(ctx, job.EnvironmentID)
	if err != nil {
		return fmt.Errorf("get environment: %w", err)
	}

	if env.ContainerID == nil {
		return fmt.Errorf("no container to start")
	}

	if err := w.docker.StartContainer(ctx, *env.ContainerID); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	if err := w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "RUNNING"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	log.Printf("Environment %s started", job.EnvironmentID)
	return nil
}

func (w *Worker) handleStop(ctx context.Context, job Job) error {
	env, err := w.queries.GetEnvironmentByID(ctx, job.EnvironmentID)
	if err != nil {
		return fmt.Errorf("get environment: %w", err)
	}

	if env.ContainerID == nil {
		return fmt.Errorf("no container to stop")
	}

	if err := w.docker.StopContainer(ctx, *env.ContainerID, 30); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	if err := w.queries.UpdateEnvironmentStatus(ctx, job.EnvironmentID, "STOPPED"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	log.Printf("Environment %s stopped", job.EnvironmentID)
	return nil
}

func (w *Worker) handleResume(ctx context.Context, job Job) error {
	return w.handleStart(ctx, job)
}

func (w *Worker) handleDelete(ctx context.Context, job Job) error {
	env, err := w.queries.GetEnvironmentByID(ctx, job.EnvironmentID)
	if err != nil {
		return fmt.Errorf("get environment: %w", err)
	}

	if env.ContainerID != nil {
		if err := w.docker.RemoveContainer(ctx, *env.ContainerID); err != nil {
			log.Printf("Failed to remove container: %v", err)
		}
	}

	if err := w.queries.SoftDeleteEnvironment(ctx, job.EnvironmentID); err != nil {
		return fmt.Errorf("soft delete environment: %w", err)
	}

	log.Printf("Environment %s deleted", job.EnvironmentID)
	return nil
}

func (w *Worker) Shutdown() {
	close(w.shutdownCh)
	w.nc.Close()
	database.Close(w.db)
}

func RunWorker() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	w, err := NewWorker(cfg)
	if err != nil {
		log.Fatalf("failed to create worker: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down worker...")
		w.Shutdown()
	}()

	ctx := context.Background()
	if err := w.Start(ctx); err != nil {
		log.Fatalf("worker error: %v", err)
	}
}