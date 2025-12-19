package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// 1. load SHA of lua script
// 2. feature: Allow

type HybridLimiter struct {
	Rdb *redis.Client
	SHA string
}

func LoadHybridLimiter(rdb *redis.Client, path string) (*HybridLimiter, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lua: %w", err)
	}

	sha := calcSHA(b)
	exists, err := rdb.ScriptExists(context.Background(), sha).Result()
	if err != nil {
		return nil, fmt.Errorf("script not exists: %w", err)
	}

	if err == nil && len(exists) > 0 && exists[0] {
		return &HybridLimiter{
			Rdb: rdb,
			SHA: sha,
		}, nil
	}
	newSha, err := rdb.ScriptLoad(context.Background(), string(b)).Result() // use complete script content
	if err != nil {
		return nil, fmt.Errorf("script not loaded: %w", err)
	}

	return &HybridLimiter{
		Rdb: rdb,
		SHA: newSha,
	}, nil
}

// Allow runs lua script to check if request is allowed
// keyPrefix is like user format => user:U1, ip format => ip:1.2.3.4
// slidingLimit==0 to disable sliding window check
func (h *HybridLimiter) Allow(ctx context.Context, keyPrefix string, maxTokens float64, refillRate float64, slidingLimit int64, slidingWindowSec int64) (allowed bool, reason string, tokensLeft float64, swCount int64, err error) {
	now := time.Now().UnixNano() / 1e9
	res, err := h.Rdb.EvalSha(ctx, h.SHA,
		[]string{
			"tb:" + keyPrefix,
			"sw:" + keyPrefix,
		},
		strconv.FormatFloat(maxTokens, 'f', -1, 64),
		strconv.FormatFloat(refillRate, 'f', -1, 64),
		fmt.Sprintf("%.6f", now),
		"1", // cost
		strconv.FormatInt(slidingLimit, 10),
		strconv.FormatInt(slidingWindowSec, 10),
	).Result()
	if err != nil {
		return false, "", 0, 0, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 4 {
		return false, "", 0, 0, fmt.Errorf("unexpected lua result: %v", res)
	}

	allowedInt := int64(0)
	switch v := arr[0].(type) {
	case int64:
		allowedInt = v
	case string:
		allowedInt, _ = strconv.ParseInt(v, 10, 64)
	}
	reason = fmt.Sprint(arr[1])
	tokensLeft, _ = strconv.ParseFloat(fmt.Sprint(arr[2]), 64)
	swCount, _ = strconv.ParseInt(fmt.Sprint(arr[3]), 10, 64)
	return allowedInt == 1, reason, tokensLeft, swCount, nil

}
