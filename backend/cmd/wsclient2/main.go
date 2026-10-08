package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNGFkMTg4YTgtNGZhYS00ZDA3LWE1MGUtZmNiODAwODc2Y2ZjIiwidXNlcm5hbWUiOiJjb2xsYWJvcmF0b3IiLCJzdWIiOiI0YWQxODhhOC00ZmFhLTRkMDctYTUwZS1mY2I4MDA4NzZjZmMiLCJleHAiOjE3OTE0NTk0NTMsImlhdCI6MTc5MTM3MzA1M30.-jeHrIgZbIdDsLw4a9GbPmco2TvY6V0pZTgVPPwbscY"

	url := "ws://localhost:8086/api/v1/environments/c4086399-ba43-44ce-8d9c-784ace869f0a/presence/ws"

	headers := map[string][]string{
		"Authorization": {"Bearer " + token},
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Collaborator connected to WebSocket")

	// Read messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var msg map[string]interface{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			fmt.Printf("Collaborator received: %+v\n", msg)
		}
	}()

	// Send ping every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
			fmt.Println("Collaborator sent ping")
		}
	}
}