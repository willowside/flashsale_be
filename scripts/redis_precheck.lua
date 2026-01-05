-- 1. Check stock key
-- 2. Check if stock > 0
-- 3. Check if user has purchased
-- 4. Success: deduct 1 stock

-- KEY[1] stock_key
-- KEY[2] user_set_key
-- ARGV[1] user_id


local stock_key = KEYS[1]
local user_set_key = KEYS[2]
local user_id = ARGV[1]


-- Get remaining stock
local stock = tonumber(redis.call("GET", stock_key))
if not stock then
    return {0, "STOCK_NOT_FOUND"}
end

if stock <= 0 then
    return {0, "OUT_OF_STOCK"}
end

-- Check if purchased
local exists = redis.call("SISMEMBER", user_set_key, user_id) -- check if user_id exists in the set held by key:user_set_key
if exists == 1 then
    return {0, "USER_ALREADY_PURCHASED"}
end

-- -- Deduct to Preserve stock
-- redis.call("DECR", stock_key)

-- Success callback
return {1, "OK"}