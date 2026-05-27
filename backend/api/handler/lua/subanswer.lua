local status = redis.call("HGET", KEYS[1], "status")
if not status then
    return 1
end

if status ~= ARGV[1] then
    return 2
end
redis.call("HSET", KEYS[1], "updated_at", ARGV[2])
redis.call("HINCRBY",KEYS[1], "current_index", 1)
redis.call("EXPIRE", KEYS[1], ARGV[3])
return 0
