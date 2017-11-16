from flask import request, redirect, Response
import redis, json

# TODO extract hardcoded location
redisConnection = redis.StrictRedis(host='redis.readitlater', port=6379, db=0)

# List all values for a given key
def list():
    key = get_key()
    bytevals = redisConnection.lrange(key, 0, -1)
    result = []
    for val in bytevals:
        result.append(val.decode('utf-8'))
    return Response(json.dumps(result), mimetype='application/json')

# Append value to values of the given key
def append():
    key = get_key()
    value = request.data
    result = redisConnection.rpush(key, value)
    return Response(json.dumps(result), mimetype='application/json')

# Set the value for a given key
def set():
    key = get_key()
    value = request.data
    result = redisConnection.set(key, value)
    return Response(json.dumps(result), mimetype='application/json')

# Get the value for a given key
def get():
    key = get_key()
    val = redisConnection.get(key)
    result = val.decode('utf-8')
    return Response(json.dumps(result), mimetype='application/json')


def get_key():
    key = request.args.get('key')
    if key is None:
        key = request.headers["X-Redis-Key"]
    return key