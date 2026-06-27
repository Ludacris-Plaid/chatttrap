package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const TokenRate = 0.50
const MinTokens = 1

type TokenService struct {
	rdb  *redis.Client
	pool *pgxpool.Pool
}

func NewTokenService(rdb *redis.Client) *TokenService {
	return &TokenService{rdb: rdb}
}

func (s *TokenService) SetPool(pool *pgxpool.Pool) {
	s.pool = pool
}

func (s *TokenService) CalculateCost(durationSeconds int) (tokens int, usd float64) {
	minutes := durationSeconds / 60
	if durationSeconds%60 > 0 {
		minutes++
	}
	if minutes < MinTokens {
		minutes = MinTokens
	}
	usd = float64(minutes) * TokenRate
	return minutes, usd
}

func (s *TokenService) DeductTokens(ctx context.Context, userID string, tokens int) error {
	if s.rdb != nil {
		lockKey := fmt.Sprintf("token_lock:%s", userID)
		locked, err := s.rdb.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
		if err != nil {
			return fmt.Errorf("redis error: %w", err)
		}
		if !locked {
			return fmt.Errorf("concurrent deduction in progress")
		}
		defer s.rdb.Del(ctx, lockKey)
	}

	// Deduct from DB balance
	if s.pool != nil {
		_, err := s.pool.Exec(ctx,
			`UPDATE profiles SET balance = GREATEST(balance - $1, 0), tokens_used = tokens_used + $2 WHERE id = $3`,
			float64(tokens)*TokenRate, tokens, userID)
		if err != nil {
			return fmt.Errorf("balance deduction failed: %w", err)
		}
	}

	return nil
}

func (s *TokenService) CanUserAfford(ctx context.Context, balance float64, tokens int) bool {
	cost := float64(tokens) * TokenRate
	return balance >= cost
}
