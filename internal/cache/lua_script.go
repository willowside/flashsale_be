package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type LuaScript struct {
	Script string
	SHA    string
}

func NewLuaScript(script string) *LuaScript {
	return &LuaScript{Script: script, SHA: script}
}

func (l *LuaScript) Run(ctx context.Context, rdb *redis.Client, keys []string, args ...any) *redis.Cmd {

	cmd := rdb.EvalSha(ctx, l.SHA, keys, args...)
	if err := cmd.Err(); err != nil && err.Error() == "NOSCRIPT No matching script. Please use EVAL." {

		// Fallback to Eval (load script automatically)
		cmd = rdb.Eval(ctx, l.Script, keys, args...)
	}

	return cmd
}
