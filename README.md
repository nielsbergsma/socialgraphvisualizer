# Social graph visualizer

##Start service##
go run main.go --neo4j="http://@localhost:7474/db/data" --twitter-consumer-key="..." --twitter-consumer-secret="..." --twitter-accesstoken="..." --twitter-accesstoken-secret="..."

##Crawl new graph##
curl -X POST -H "Content-Type: application/json" -d '{ "uri" : "twitter://user/371238667" }' "http://localhost:7999/graphs"

##View data##
Open browser: http://localhost:7999/graphs/dHdpdHRlcjovL3VzZXIvMzcxMjM4NjY3
 - /graphs/(id) = base64 encoded: twitter://user/(id)

##View graph##
Open browser: http://localhost:7999/

![alt text](https://github.com/nielsbergsma/socialgraphvisualizer/notes/example.png "Example lay-out")
