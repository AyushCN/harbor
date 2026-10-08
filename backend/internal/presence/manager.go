package presence

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	presenceChannel = "presence:events"
	presencePrefix  = "presence:environment:"
)

type PresenceEvent struct {
	Type      string    `json:"type"` // "join", "leave", "sync"
	EnvironmentID string  `json:"environment_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}

type UserPresence struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Manager struct {
	client    *redis.Client
	pubsub    *redis.PubSub
	ctx       context.Context
	cancel    context.CancelFunc
	handlers  map[string][]chan PresenceEvent // environment_id -> channels
}

func NewManager(client *redis.Client) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string][]chan PresenceEvent),
	}
	
	// Subscribe to presence events
	m.pubsub = m.client.Subscribe(ctx, presenceChannel)
	go m.listen()
	
	return m
}

func (m *Manager) listen() {
	ch := m.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			var event PresenceEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to unmarshal presence event: %v", err)
				continue
			}
			m.broadcast(event)
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) broadcast(event PresenceEvent) {
	if channels, ok := m.handlers[event.EnvironmentID]; ok {
		for _, ch := range channels {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

func (m *Manager) Subscribe(environmentID string) chan PresenceEvent {
	ch := make(chan PresenceEvent, 10)
	m.handlers[environmentID] = append(m.handlers[environmentID], ch)
	return ch
}

func (m *Manager) Unsubscribe(environmentID string, ch chan PresenceEvent) {
	if channels, ok := m.handlers[environmentID]; ok {
		for i, c := range channels {
			if c == ch {
				m.handlers[environmentID] = append(channels[:i], channels[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

func (m *Manager) Join(ctx context.Context, environmentID, userID, username, avatarURL, role string) error {
	key := presencePrefix + environmentID
	userData := UserPresence{
		UserID:    userID,
		Username:  username,
		AvatarURL: avatarURL,
		Role:      role,
		JoinedAt:  time.Now(),
	}
	
	jsonData, err := json.Marshal(userData)
	if err != nil {
		return err
	}
	
	// Add to Redis set
	if err := m.client.HSet(ctx, key, userID, jsonData).Err(); err != nil {
		return err
	}
	
	// Publish join event
	event := PresenceEvent{
		Type:           "join",
		EnvironmentID:  environmentID,
		UserID:         userID,
		Username:       username,
		AvatarURL:      avatarURL,
		Role:           role,
		Timestamp:      time.Now(),
	}
	
	eventJSON, _ := json.Marshal(event)
	return m.client.Publish(ctx, presenceChannel, eventJSON).Err()
}

func (m *Manager) Leave(ctx context.Context, environmentID, userID string) error {
	key := presencePrefix + environmentID
	
	// Remove from Redis set
	if err := m.client.HDel(ctx, key, userID).Err(); err != nil {
		return err
	}
	
	// Publish leave event
	event := PresenceEvent{
		Type:          "leave",
		EnvironmentID: environmentID,
		UserID:        userID,
		Timestamp:     time.Now(),
	}
	
	eventJSON, _ := json.Marshal(event)
	return m.client.Publish(ctx, presenceChannel, eventJSON).Err()
}

func (m *Manager) GetPresence(ctx context.Context, environmentID string) ([]UserPresence, error) {
	key := presencePrefix + environmentID
	
	data, err := m.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	var users []UserPresence
	for _, jsonData := range data {
		var user UserPresence
		if err := json.Unmarshal([]byte(jsonData), &user); err != nil {
			log.Printf("Failed to unmarshal user presence: %v", err)
			continue
		}
		users = append(users, user)
	}
	
	return users, nil
}

func (m *Manager) Close() {
	m.cancel()
	if m.pubsub != nil {
		m.pubsub.Close()
	}
}