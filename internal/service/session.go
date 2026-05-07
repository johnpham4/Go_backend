package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

const sessionTTL = 24 * time.Hour

type SessionService struct {
	rdb *redis.Client
}

func NewSessionService(rdb *redis.Client) *SessionService {
	return &SessionService{rdb: rdb}
}

func (s *SessionService) CreateSession(ctx context.Context, username string) (string, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return "", err
	}

	key := sessionKey(sessionID)
	if err := s.rdb.Set(ctx, key, username, sessionTTL).Err(); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *SessionService) GetUsername(ctx context.Context, sessionID string) (string, error) {
	key := sessionKey(sessionID)
	return s.rdb.Get(ctx, key).Result()
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}