package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZDIyNzY5NDEtNzIwMy00YTMyLTlmYTQtZDliNDUxMjVhYjI4IiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsInN1YiI6ImQyMjc2OTQxLTcyMDMtNGEzMi05ZmE0LWQ5YjQ1MTI1YWIyOCIsImV4cCI6MTc5MTQ1ODgxNiwiaWF0IjoxNzkxMzcyNDE2fQ.fq3YY5vPxzCPXAxF68PmxN_PXsmClEF_3wLd0zMAZfA"

	url := "ws://localhost:8086/api/v1/environments/c4086399-ba43-44ce-8d9c-784ace869f0a/presence/ws"

	headers := map[string][]string{
		"Authorization": {"Bearer " + token},
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket")

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
			fmt.Printf("Received: %+v\n", msg)
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
			fmt.Println("Sent ping")
		}
	}
}