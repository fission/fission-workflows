import opengraph
from flask import request, current_app

def main():
    doc = request.data
    ogp = opengraph.OpenGraph(html=doc)
    return ogp.to_json()