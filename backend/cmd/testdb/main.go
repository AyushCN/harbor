package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/AyushCN/harbor/internal/config"
	"github.com/AyushCN/harbor/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close(pool)

	queries := database.NewQueries(pool)

	userID := uuid.MustParse("d2276941-7203-4a32-9fa4-d9b45125ab28")
	
	workspace, err := queries.CreateWorkspace(ctx, userID, "https://github.com/test/repo2", "main", "my-app-2")
	if err != nil {
		log.Fatalf("CreateWorkspace failed: %v", err)
	}
	fmt.Printf("Workspace created: %+v\n", workspace)

	env, err := queries.CreateEnvironment(ctx, workspace.ID, userID)
	if err != nil {
		log.Fatalf("CreateEnvironment failed: %v", err)
	}
	fmt.Printf("Environment created: %+v\n", env)

	member, err := queries.AddEnvironmentMember(ctx, env.ID, userID, "OWNER", nil)
	if err != nil {
		log.Fatalf("AddEnvironmentMember failed: %v", err)
	}
	fmt.Printf("Member added: %+v\n", member)
}