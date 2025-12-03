-- worker calls the script after successful purchased
-- finalize: add user to purchased set

-- KEYS[1] = user_set_key
-- ARGV[1] = user_id

local user_set_key = KEYS[1]
local user_id = ARGV[1]

redis.call("SADD", user_set_key, user_id) -- add user_id into set held by key:user_set_key
return 1