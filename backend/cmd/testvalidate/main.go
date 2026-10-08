package main

import (
	"fmt"
	"github.com/yourusername/harbor/internal/auth"
	"github.com/yourusername/harbor/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDIyNzY5NDEtNzIwMy00YTMyLTlmYTQtZDliNDUxMjVhYjI4IiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsInN1YiI6ImQyMjc2OTQxLTcyMDMtNGEzMi05ZmE0LWQ5YjQ1MTI1YWIyOCIsImV4cCI6MTc5MTQ1ODgxNiwiaWF0IjoxNzkxMzcyNDE2fQ.fq3YY5vPxzCPXAxF68PmxN_PXsmClEF_3wLd0zMAZfA"

	claims, err := auth.ValidateToken(token, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("UserID: %s, Username: %s\n", claims.UserID, claims.Username)
}