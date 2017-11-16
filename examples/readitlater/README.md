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
save-article | workflow | Given an url, process page and store resulting article
parse-article | workflow | Given a html document, parse and return parsed article
DONE http     | binary   | Perform an HTTP request and return response
DONE parse-article-body | python | Parse article body from html document
extract-ogp | python | Given HTML document, parse ogp data from it  
DONE notify-pushbullet | python | Send notification to pushbullet
DONE redis-append | python | Stores item into redis
DONE redis-list | python| Lists items for given key


## TODO
- Some sort of config/secret management
- Implement image handling/transformation
- Show off some loop construct
- 
