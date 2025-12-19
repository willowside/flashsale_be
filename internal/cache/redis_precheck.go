package cache

import (
	"context"
	"errors"
	"fmt"
)

var LuaSHAScripts *LuaScripts // store loaded script SHA globally

type PreCheckResult struct {
	Success bool
	Reason  string
}

// Called by main.go after loading scripts
func SetLuaScripts(s *LuaScripts) {
	LuaSHAScripts = s
}

func FlashSalePreCheck(productID, userID string) (*PreCheckResult, error) {
	if Rdb == nil {
		return nil, errors.New("redis not init")
	}
	if LuaSHAScripts == nil || LuaSHAScripts.PrecheckSHA.SHA == "" {
		return nil, errors.New("lua scripts not loaded")
	}

	stockKey := fmt.Sprintf("flashsale:stock:%s", productID)
	// store purchasers set per product
	userSetKey := fmt.Sprintf("flashsale:purchased:%s", productID)

	// Use SHA instead of raw Lua file
	// EvalSha : KEYS=[stockKey, userSetKey], ARGV=[userID(,ttlSeconds)]
	ctx := context.Background()
	res, err := Rdb.EvalSha(ctx,
		LuaSHAScripts.PrecheckSHA.SHA,
		[]string{stockKey, userSetKey},
		userID,
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
