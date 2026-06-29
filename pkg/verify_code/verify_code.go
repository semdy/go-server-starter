package verify_code

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
)

// Type represents the purpose of a verification code.
type Type string

const (
	SMSTypeLogin   Type = "sms:login"
	EmailTypeLogin Type = "email:login"
)

// Store manages verification code lifecycle in Redis.
type Store struct {
	client         *redis.Client
	codeTTL        time.Duration
	resendCooldown time.Duration
}

// NewStore creates a verification code store.
func NewStore(client *redis.Client, ttl, cooldown time.Duration) *Store {
	return &Store{client: client, codeTTL: ttl, resendCooldown: cooldown}
}

// key returns the Redis key for a given type and identifier.
func (s *Store) key(typ Type, target string) string {
	return fmt.Sprintf("verify:%s:%s", typ, target)
}

// cooldownKey returns the Redis key for resend rate limiting.
func (s *Store) cooldownKey(typ Type, target string) string {
	return fmt.Sprintf("verify:cooldown:%s:%s", typ, target)
}

// Generate creates a 6-digit code, stores it in Redis, and returns the code.
// Returns an error if the cooldown period has not elapsed.
func (s *Store) Generate(ctx context.Context, typ Type, target string) (string, error) {
	// Check resend cooldown
	cdKey := s.cooldownKey(typ, target)
	exists, err := s.client.Exists(ctx, cdKey).Result()
	if err != nil {
		return "", fmt.Errorf("check cooldown: %w", err)
	}
	if exists > 0 {
		return "", ErrResendCooldown
	}

	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	key := s.key(typ, target)

	if err := s.client.Set(ctx, key, code, s.codeTTL).Err(); err != nil {
		return "", fmt.Errorf("store code: %w", err)
	}
	// Set cooldown
	if err := s.client.Set(ctx, cdKey, "1", s.resendCooldown).Err(); err != nil {
		return "", fmt.Errorf("set cooldown: %w", err)
	}

	return code, nil
}

// Validate checks the code for the given type and target. Returns nil on success.
// The code is deleted after successful validation.
func (s *Store) Validate(ctx context.Context, typ Type, target, code string) error {
	key := s.key(typ, target)
	stored, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return ErrCodeExpired
	}
	if err != nil {
		return fmt.Errorf("get code: %w", err)
	}
	if stored != code {
		return ErrCodeMismatch
	}
	// Delete the code so it can't be reused
	s.client.Del(ctx, key)
	return nil
}

// ---- errors ----

var (
	ErrCodeExpired     = errors.New("verification code expired")
	ErrCodeMismatch    = errors.New("verification code incorrect")
	ErrResendCooldown  = errors.New("please wait before requesting another code")
)
