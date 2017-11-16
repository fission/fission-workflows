from flask import request, redirect
import redis, json

# TODO extract hardcoded location
redisConnection = redis.StrictRedis(host='redis.readitlater', port=6379, db=0)

# List all values for a given key
def list():
    key = request.args.get('key')
    vals = redisConnection.lrange(key, 0, -1)
    return json.dumps(vals)

# Append value to values of the given key
def append():
    key = request.args.get('key')
    value = request.datas
    result = redisConnection.rpush(key, value)
    return json.dumps(result)

# Set the value for a given key
def set():
    key = request.args.get('key')
    value = request.data
    result = redisConnection.set(key, value)
    return json.dumps(result)

# Get the value for a given key
def get():
    key = request.args.get('key')
    val = redisConnection.get(key)
    return json.dumps(val)