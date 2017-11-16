from opengraph import OpenGraph
from flask import request, current_app, Response
import json

def extract():
    doc = request.get_data(as_text=True)
    if doc is None or len(doc) == 0:
        return json.dumps({})
    og = OpenGraph(html=doc)
    return Response(json.dumps(og.__data__), mimetype='application/json')