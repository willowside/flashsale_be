package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

type TokenBucketLimiter struct {
	SHA string
}

func LoadTokenBucketLua(path string) (*TokenBucketLimiter, error) {
	// Read lua file
	bucketScript, err := os.ReadFile(path)
	// error load
	if err != nil {
		return nil, fmt.Errorf("error read token bucket lua: %w", err)
	}

	sha, err := Rdb.ScriptLoad(context.Background(), string(bucketScript)).Result()
	if err != nil {
		return nil, fmt.Errorf("error load token bucket lua to redis: %w", err)
	}
	return &TokenBucketLimiter{
		SHA: sha,
	}, nil
}

// Evaluate token bucket
// KEYS[1] = bucket key
// ARGV[1] = capacity / max_tokens
// ARGV[2] = refill_rate (tokens per second)

// key prefix: "ratelimit:user:{userID}"
// maxTokens = bucket capacity
// refillRate = tokens added per second
func (tbl *TokenBucketLimiter) Allow(keyPrefix string, maxTokens, refillRate float64) (allowed bool, remaining float64, err error) {
	now := time.Now().UnixNano() / 1e9 // current time in secons

	res, err := Rdb.EvalSha(
		context.Background(),
		tbl.SHA,
		[]string{keyPrefix},
		maxTokens,
		refillRate,
		now,
		1, // consume 1 token
	).Result()
	if err != nil {
		return false, 0, fmt.Errorf("error evaluate token bucket lua: %w", err)
	}

	arr := res.([]interface{})
	allowed = arr[0].(int64) == 1
	remaining, _ = strconv.ParseFloat(fmt.Sprint(arr[1]), 64)
	return allowed, remaining, nil
}
