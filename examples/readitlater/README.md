# ReadItLater

This comprehensive example shows the main features of the workflow system. 
ReadItLater is a serverless version of a bookmarking system, 
akin to [Pocket](https://getpocket.com/) or [Wallabag](https://wallabag.org/en).

It features:
- bookmarking of an url
- parsing of bookmarked pages to readable articles
- notifying the user of the status of saved articles
- storing parsed articles into a persistent database

## Functions

Name     | Language | Description
---------|----------|------------
DONE http     | binary   | Perform an HTTP request and return response
save-article | workflow | Given an url, process page and store resulting article
parse-article | workflow | Given a html document, parse and return parsed article
parse-article-body | python | Parse article body from html document
parse-article-metadata | python | Parse article metadata (author, date, image) from html document
parse-article-thumbmail | ??? |
DONE extract-ogp | python | Given HTML document, parse ogp data from it  
resize-image | binary | Resizes an image
DONE notify-pushbullet | python | Send notification to pushbullet
store-redis | ??? | Stores json into database


## TODO
- Some sort of config/secret management
