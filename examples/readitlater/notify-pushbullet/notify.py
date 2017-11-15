from pushbullet import Pushbullet
from flask import request, current_app

HEADER_PB_APIKEY = "X-Pushbullet-ApiKey"

def main():

    try:
        api_key = request.headers[HEADER_PB_APIKEY]
    except KeyError:
        return f"No Pushbullet Apikey provided (expected in header: '{HEADER_PB_APIKEY}'", 400

    data = request.json
    note_title = data["title"]
    note_body = data["body"]

    current_app.logger.info(f"Sending pushbullet notification: '{note_title} -- {note_body}'")
    pb = Pushbullet(api_key)
    push = pb.push_note(note_title, note_body)
    return push

