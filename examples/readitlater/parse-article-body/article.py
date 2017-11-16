from flask import request, current_app, Response
from bs4 import BeautifulSoup
import json

def main():
    doc = request.get_data()
    current_app.logger.info(doc)
    soup = BeautifulSoup(doc, 'html.parser')

    # Simple text extractor
    article_nodes = []
    for node in soup(['p','h1','h2','h3','h4','h5','h6']):
        article_nodes.append(node.get_text())

    return Response(json.dumps(article_nodes), mimetype='application/json')
