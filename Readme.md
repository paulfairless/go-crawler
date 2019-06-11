# Golang Crawler #


### Dependencies ###
`go get github.com/PuerkitoBio/goquery`

`go get github.com/jarcoal/httpmock`

### Run ###
`go run crawler.go -url=<seedUrl> -depth=<crawl depth>` 

## Testing ##
`go test`

## basic crawler implementation ##
A crawl is implemented using Depth-first recursion, the main reason to do this is we can control the depth of the search and better manage the resources used.
```initialize Queue
enqueue seed

while Queue is not empty:
    URL = Pop element from Queue
    Page = Visit(URL)
    Links = ExtractLinks(Page)
    Enqueue Links on Queue
```

### Enhancements ###
* Be considerate, throttle crawl and honor Robots.txt
* goquery can be slow for larger documents, a streaming implementation would be more memory efficient
* set GOMAXPROCS to improve performance
* explore using worker groups 
* remove depth and introduce loop protection
* better error handling
* test coverage as only happy path covered
* ignore url fragments
* implement set to remove duplicated links like in header and footer
