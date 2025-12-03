package cache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/redis/go-redis/v9"
)

type LuaScripts struct {
	PrecheckSHA string
	FinalizeSHA string
}

func LoadLuaScripts(rdb *redis.Client, scriptDir string) (*LuaScripts, error) {
	ctx := context.Background()

	precheckPath := filepath.Join(scriptDir, "redis_precheck.lua")
	finalizePath := filepath.Join(scriptDir, "precheck_final.lua")

	precheck, err := loadAndRegister(ctx, rdb, precheckPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load precheck lua: %w", &err)
	}

	finalize, err := loadAndRegister(ctx, rdb, finalizePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load finalize lua: %w", &err)
	}

	return &LuaScripts{
		PrecheckSHA: precheck,
		FinalizeSHA: finalize,
	}, nil

}

func loadAndRegister(ctx context.Context, rdb *redis.Client, path string) (string, error) {

	// 1. read file
	code, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read script: %w", err)
	}

	// 2. calculate SHA
	sha := calcSHA(code)

	// don't load if SHA exists in Redis
	exists := rdb.ScriptExists(ctx, sha).Val()
	if len(exists) > 0 && exists[0] {
		return sha, nil
	}

	// 3. upload to redis
	newSHA, err := rdb.ScriptLoad(ctx, string(code)).Result()
	if err != nil {
		return "", fmt.Errorf("script load failed: %w", err)
	}

	return newSHA, nil
}

func calcSHA(code []byte) string {
	sum := sha1.Sum(code)
	return hex.EncodeToString(sum[:])
}
