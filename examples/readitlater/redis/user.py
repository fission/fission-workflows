from flask import request, redirect
import redis, json

# TODO extract hardcoded location
redisConnection = redis.StrictRedis(host='redis.readitlater', port=6379, db=0)

# List all values for a given key
def list():
    key = get_key()
    bytevals = redisConnection.lrange(key, 0, -1)
    li = []
    for val in bytevals:
        li.append(val.decode('utf-8'))
    return json.dumps(li)

# Append value to values of the given key
def append():
    key = get_key()
    value = request.data
    result = redisConnection.rpush(key, value)
    return json.dumps(result)

# Set the value for a given key
def set():
    key = get_key()
    value = request.data
    result = redisConnection.set(key, value)
    return json.dumps(result)

# Get the value for a given key
def get():
    key = get_key()
    val = redisConnection.get(key)
    return json.dumps(val.decode('utf-8'))


def get_key():
    key = request.args.get('key')
    if key is None:
        key = request.headers["X-Redis-Key"]
    return key