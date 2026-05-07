package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pingLockKey      = "lock:ping"
	pingLockTTL      = 5 * time.Second
	pingLeaderboard  = "ping:leaderboard"
	pingHLLKey       = "ping:hll"
	rateLimitWindow  = 60 * time.Second
	rateLimitMaxHits = 2
)

var releaseLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

type PingService struct {
	rdb *redis.Client
}

func NewPingService(rdb *redis.Client) *PingService {
	return &PingService{rdb: rdb}
}

func (s *PingService) AcquireLock(ctx context.Context) (string, bool, error) {
	token, err := newToken()
	if err != nil {
		return "", false, err
	}

	ok, err := s.rdb.SetNX(ctx, pingLockKey, token, pingLockTTL).Result()
	return token, ok, err
}

func (s *PingService) ReleaseLock(ctx context.Context, token string) error {
	_, err := releaseLockScript.Run(ctx, s.rdb, []string{pingLockKey}, token).Result()
	return err
}

func (s *PingService) CheckRateLimit(ctx context.Context, username string) (bool, int64, error) {
	key := rateLimitKey(username)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	if count == 1 {
		if err := s.rdb.Expire(ctx, key, rateLimitWindow).Err(); err != nil {
			return false, 0, err
		}
	}

	if count > rateLimitMaxHits {
		return false, count, nil
	}

	return true, count, nil
}

func (s *PingService) IncrementCount(ctx context.Context, username string) (int64, error) {
	key := countKey(username)
	return s.rdb.Incr(ctx, key).Result()
}

func (s *PingService) UpdateStats(ctx context.Context, username string) error {
	pipe := s.rdb.TxPipeline()
	pipe.ZIncrBy(ctx, pingLeaderboard, 1, username)
	pipe.PFAdd(ctx, pingHLLKey, username)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *PingService) GetTop(ctx context.Context, limit int64) ([]redis.Z, error) {
	if limit <= 0 {
		return []redis.Z{}, nil
	}

	return s.rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   pingLeaderboard,
		Start: 0,
		Stop:  limit - 1,
		Rev:   true,
	}).Result()
}

func (s *PingService) GetUniqueCount(ctx context.Context) (int64, error) {
	return s.rdb.PFCount(ctx, pingHLLKey).Result()
}

func rateLimitKey(username string) string {
	return "ping:rl:" + username
}

func countKey(username string) string {
	return "ping:count:" + username
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
