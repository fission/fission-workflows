from opengraph import OpenGraph
from flask import request, current_app
import json

def main():
    doc = request.data(as_text=True)
    og = OpenGraph(html=doc)
    return json.dumps(og)