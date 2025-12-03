package cache

import (
	"context"
	"errors"
	"fmt"
)

var luaScripts *LuaScripts // store loaded script SHA globally

type PreCheckResult struct {
	Success bool
	Reason  string
}

// Called by main.go after loading scripts
func SetLuaScripts(s *LuaScripts) {
	luaScripts = s
}

func FlashSalePreCheck(productID, userID string) (*PreCheckResult, error) {
	if Rdb == nil {
		return nil, errors.New("redis not init")
	}
	if luaScripts == nil || luaScripts.PrecheckSHA == "" {
		return nil, errors.New("lua scripts not loaded")
	}

	stockKey := fmt.Sprintf("flashsale:stock:%s", productID)
	// store purchasers set per product
	userSetKey := fmt.Sprintf("flashsale:users:%s", productID)

	// Use SHA instead of raw Lua file
	// EvalSha : KEYS=[stockKey, userSetKey], ARGV=[userID, ttlSeconds(optional)]
	ctx := context.Background()
	res, err := Rdb.EvalSha(ctx,
		luaScripts.PrecheckSHA,
		[]string{stockKey, userSetKey},
		userID, 60, productID,
	).Result()
	if err != nil {
		return nil, err
	}

	// Lua returns {code, reason}
	data, ok := res.([]interface{})
	if !ok || len(data) < 2 {
		return nil, fmt.Errorf("unexpected lua retuen: %#v", res)
	}
	code, ok := data[0].(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected code type from lua: %T", data[0])
	}
	reason, _ := data[1].(string)

	return &PreCheckResult{
		Success: code == 1,
		Reason:  reason,
	}, nil
}
