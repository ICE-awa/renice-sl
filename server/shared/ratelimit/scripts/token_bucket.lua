local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])

local now = redis.call("TIME")
local now_ms = now[1] * 1000 + math.floor(now[2] / 1000)

local bucket = redis.call("HMGET", key, "tokens", "updated_at")
local tokens = tonumber(bucket[1])
local updated_at = tonumber(bucket[2])

-- 如果 tokens 不存在则初始化
if tokens == nil then
    tokens = capacity
    updated_at = now_ms
end

local duration = math.max(0, now_ms - updated_at)
local refill = refill_rate * duration / 1000
tokens = math.min(capacity, tokens + refill)

local allowed = 0
if tokens >= cost then
   tokens = tokens - cost
   allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "updated_at", now_ms)
redis.call("PEXPIRE", key, math.max(60000, math.ceil(capacity / refill_rate * 2000)))

return allowed