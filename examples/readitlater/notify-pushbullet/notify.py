from pushbullet import Pushbullet
from pushbullet.errors import InvalidKeyError
from flask import request, current_app, Response
import json

HEADER_PB_APIKEY = "X-Pushbullet-ApiKey"

def main():
    try:
        api_key = request.headers[HEADER_PB_APIKEY]
    except KeyError:
        return "No Pushbullet Apikey provided (expected in header: '{}')".format(HEADER_PB_APIKEY), 400

    payload = json.loads(request.get_data(as_text=True))
    current_app.logger.info(payload)
    note_title = payload["title"]
    note_body = payload["body"]

    current_app.logger.info("Sending pushbullet notification: '{} -- {}'".format(note_title, note_body))
    try:
        pb = Pushbullet(api_key)
    except InvalidKeyError:
        return "Provided Pushbullet Apikey is invalid", 400
    push = pb.push_note(note_title, note_body)

    return Response(json.dumps(push), mimetype='application/json')
