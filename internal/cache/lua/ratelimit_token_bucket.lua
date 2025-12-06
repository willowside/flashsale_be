-- KEYS[1] = bucket key
-- ARGV[1] = capacity / max_tokens
-- ARGV[2] = refill_rate (tokens per second)
-- ARGV[3] = now (current unix timestamp as float)
-- ARGV[4] = cost (1)

local bucket_key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

-- read stored bucket
-- returns a Lua table containing the values corresponding to the requested fields from a hash stored at given key
local bucket = redis.call("HMGET", bucket_key, "tokens", "last_refill")

local tokens = bucket[1]
local last_refill = bucket[2]

if tokens == false or tokens == nil then
    tokens = max_tokens
    last_refill = now
else
    tokens = tonumber(tokens)
    last_refill = tonumber(last_refill)
end

-- refill
local delta = now - last_refill
local refill = delta * refill_rate
tokens = math.min(max_tokens, tokens + refill)
last_refill = now

-- check
if tokens >= cost then
    tokens = tokens - cost
    redis.call("HMSET", bucket_key, "tokens", tokens, "last_refill", last_refill)
    redis.call("EXPIRE", bucket_key, 60) -- keep bucket alive
    return {1, tokens} -- allowed
else
    redis.call("HMSET", bucket_key, "tokens", tokens, "last_refill", last_refill)
    redis.call("EXPIRE", bucket_key, 60) -- keep bucket alive
    return {0, tokens} -- rejected
end