package chatbot

import (
	"sync"
	"time"
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
