package store

import (
	"context"
	"sync"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

type MemoryStore struct {
	mu         sync.Mutex
	challenges map[string]domain.Challenge
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{challenges: map[string]domain.Challenge{}}
}

func (s *MemoryStore) GetChallenge(_ context.Context, key string) (*domain.Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge, ok := s.challenges[key]
	if !ok {
		return nil, nil
	}
	return &challenge, nil
}

func (s *MemoryStore) PutChallenge(_ context.Context, key string, challenge domain.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.challenges[key] = challenge
	return nil
}

func (s *MemoryStore) DeleteChallenge(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.challenges, key)
	return nil
}
