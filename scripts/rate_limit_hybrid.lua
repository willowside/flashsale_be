-- hybrid rate limiter
-- First do sliding window counter (short term attacks), then token bucket (long term rate limit, burst)


-- KEYS[1] = token_bucket_key, eg. "tb:user:U1"
-- KEYS[2] = sliding_key, eg. "sl:user:U1"

-- ARGV[1] = max_tokens: capacity, integer
-- ARGV[2] = refill_rate: tokens per second, number
-- ARGV[3] = now: seconds in float
-- ARGV[4] = cost: tokens to consume, integer, usually 1
-- ARGV[5] = sliding_limit: max reqs allowed in window
-- ARGV[6] = sliding_window_seconds, window size, integer


local tb_key = KEYS[1]
local sw_key = KEYS[2]

local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local sw_limit = tonumber(ARGV[5])
local sw_size = tonumber(ARGV[6])

-- token bucket
local tb_vals = redis.call("HGET", tb_key, "tokens", "last_refill")
local tokens = tb_vals[1]
local last_refill = tb_vals[2]

if tokens == false or tokens == nil then
    tokens = max_tokens
    last_refill = now
else
    tokens = tonumber(tokens)
    last_refill = tonumber(last_refill)
end

local delta = math.max(0, now - last_refill)
local refill = delta * refill_rate
tokens = math.min(max_tokens, tokens + refill)
last_refill = now

-- simple fixed window counter
local sw_count = redis.call("INCR", sw_key)
if tonumber(sw_count) == 1 then
    redis.call("EXPIRE", sw_key, sw_size)
end

-- 1. Check sliding limit, if exceeded, update TB store & reject
if sw_limit > 0 and tonumber(sw_count) > sw_limit then
    -- persist token bucket state & return rejection with reason sliding
    redis.call("HMSET", tb_key, "tokens", tokens, "last_refill", last_refill)
    redis.call("EXPIRE", tb_key, math.ceil(math.max(1, (max_tokens / (refill_rate + 1e-9)) * 2)))
    return {0, "sliding", tokens, sw_count}
end


-- 2. Check tocken bucket, if enough tokens, consume & accept
-- return format: {allowed: 1/0, reason:"token"/"sliding", tokens_left, sliding_count}
if tokens >= cost then
    tokens = tokens - cost
    redis.call("HMSET", tb_key, "tokens", tokens, "last_refill", last_refill)
    redis.call("EXPIRE", tb_key, math.ceil(math.max(1, (max_tokens / (refill_rate + 1e-9)) * 2)))
    return {1, "token", tokens, sw_count}
else
    -- not enough tokens, persist state & reject
    redis.call("HMSET", tb_key, "tokens", tokens, "last_refill", last_refill)
    redis.call("EXPIRE", tb_key, math.ceil(math.max(1, (max_tokens / (refill_rate + 1e-9)) * 2)))
    return {0, "token", tokens, sw_count}
end
