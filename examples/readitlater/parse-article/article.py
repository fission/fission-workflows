from flask import request, current_app
from bs4 import BeautifulSoup

def main():
    doc = request.data
    soup = BeautifulSoup(doc, 'html.parser')

    # Simple text extractor
    article_nodes = []
    for node in soup(['p','h1','h2','h3','h4','h5','h6']):
        article_nodes.append(node.get_text())

    return article_nodes
