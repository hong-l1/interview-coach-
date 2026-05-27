local status = redis.call("HGET", KEYS[1], "status")
if not status then
    return 1
end

if status ~= ARGV[1] then
    return 2
end

redis.call("HSET", KEYS[1],
    "status", ARGV[2],
    "ended_at", ARGV[3],
    "updated_at", ARGV[3]
)
redis.call("EXPIRE", KEYS[1], ARGV[4])
redis.call("EXPIRE", KEYS[2], ARGV[4])
return 0
