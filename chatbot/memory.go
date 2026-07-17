package chatbot

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type StateStore interface {
	Get(chatID string) ChatState
	Save(state ChatState)
	Clear(chatID string)
}

type InMemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]ChatState
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		states: map[string]ChatState{},
	}
}

func (s *InMemoryStateStore) Get(chatID string) ChatState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.states[chatID]
	if !ok {
		return ChatState{ChatID: chatID}
	}
	return state
}

func (s *InMemoryStateStore) Save(state ChatState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state.UpdatedAt = time.Now()
	s.states[state.ChatID] = state
}

func (s *InMemoryStateStore) Clear(chatID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, chatID)
}

type RedisStateStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisStateStore(client *redis.Client) *RedisStateStore {
	return &RedisStateStore{
		client: client,
		prefix: "chatbot:state:",
		ttl:    24 * time.Hour,
	}
}

func (s *RedisStateStore) Get(chatID string) ChatState {
	var state ChatState
	value, err := s.client.Get(context.Background(), s.key(chatID)).Result()
	if err == redis.Nil {
		return ChatState{ChatID: chatID}
	}
	if err != nil {
		log.Printf("failed to get chat state from redis: %v", err)
		return ChatState{ChatID: chatID}
	}
	if err := json.Unmarshal([]byte(value), &state); err != nil {
		log.Printf("failed to decode chat state from redis: %v", err)
		return ChatState{ChatID: chatID}
	}
	if state.ChatID == "" {
		state.ChatID = chatID
	}
	return state
}

func (s *RedisStateStore) Save(state ChatState) {
	state.UpdatedAt = time.Now()
	payload, err := json.Marshal(state)
	if err != nil {
		log.Printf("failed to encode chat state for redis: %v", err)
		return
	}
	if err := s.client.Set(context.Background(), s.key(state.ChatID), payload, s.ttl).Err(); err != nil {
		log.Printf("failed to save chat state to redis: %v", err)
	}
}

func (s *RedisStateStore) Clear(chatID string) {
	if err := s.client.Del(context.Background(), s.key(chatID)).Err(); err != nil {
		log.Printf("failed to clear chat state from redis: %v", err)
	}
}

func (s *RedisStateStore) key(chatID string) string {
	return s.prefix + chatID
}
