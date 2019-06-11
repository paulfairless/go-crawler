package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

/*
	Main method to start crawl
	url - start url
	depth - depth to crawl
*/
func main() {
	seed := flag.String("url", "https://monzo.com", "url to spider")
	depth := flag.Int("depth", 2, "crawl depth")
	flag.Parse()

	startUrl, err := url.Parse(*seed)
	if err != nil {
		log.Println("Could not parse start url")
		os.Exit(1)
	}

	fetcher := CachedPageFetcher{visited: make(map[string]Page)}
	result := make(chan string)
	go Crawl(startUrl, 0, *depth, fetcher, result)

	for s := range result {
		fmt.Println(s)
	}
}

/* Process Url from channel */
func Crawl(url *url.URL, depth int, maxDepth int, fetcher Fetcher, ret chan string) {
	defer close(ret)
	spacing := strings.Repeat("\t", depth)

	if depth >= maxDepth {
		ret <- fmt.Sprintf("%s|_ %s", spacing, url)
		return
	}

	page, err := fetcher.Fetch(url)
	if err != nil {
		ret <- err.Error()
		return
	}
	ret <- fmt.Sprintf("%s|_ %s %q", spacing, url, page.title)

	result := make([]chan string, len(page.urls))
	for i, u := range page.urls {
		result[i] = make(chan string)
		go Crawl(u, depth+1, maxDepth, fetcher, result[i])
	}

	for i := range result {
		for s := range result[i] {
			ret <- s
		}
	}

	return
}

/* Data structure for page */
type Page struct {
	title string
	urls  []*url.URL
}

/* Fetcher interface */
type Fetcher interface {
	Fetch(*url.URL) (*Page, error)
}

/* Cached Page fetcher */
type CachedPageFetcher struct {
	visited map[string]Page
	mux     sync.Mutex
}

/* 	safe simple cache for fetched pages */
func (c *CachedPageFetcher) Set(s string, page Page) {
	c.mux.Lock()
	c.visited[s] = page
	c.mux.Unlock()
}
func (c *CachedPageFetcher) Get(s string) (Page, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	val, ok := c.visited[s]
	return val, ok
}

/* Parse links from page */
func (c *CachedPageFetcher) ParsePage(r io.Reader, pageUrl *url.URL) (Page, error) {

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Println(err)
		return Page{}, err
	}

	// parse page title
	title := strings.TrimSpace(doc.Find("head title").Text())
	var links []*url.URL

	doc.Find("body a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		url, _ := url.Parse(link)
		resolvedLink := pageUrl.ResolveReference(url)

		// verify links have same domain
		if pageUrl.String() != resolvedLink.String() && pageUrl.Host == resolvedLink.Host {
			links = append(links, resolvedLink)
		} else {
			// reject link on different domain
		}
	})

	return Page{title, links}, nil
}

/* Download page, parse results and save in cache */
func (c CachedPageFetcher) Fetch(pageUrl *url.URL) (*Page, error) {
	// Check cache for page
	page, ok := c.Get(pageUrl.String())
	if ok {
		return &page, nil
	}

	// Get page via http
	resp, err := http.Get(pageUrl.String())
	if err != nil {
		log.Println("Failed to fetch url:"+pageUrl.String(), err)
		return &Page{}, err
	}
	defer resp.Body.Close()
	page, err = c.ParsePage(resp.Body, pageUrl)
	if err != nil {
		log.Println("Failed to parse page", err)
		return &page, err
	}
	// Set page in cache
	c.Set(pageUrl.String(), page)

	return &page, nil
}
