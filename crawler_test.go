package main

import (
	"github.com/jarcoal/httpmock"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

var fetcher = CachedPageFetcher{visited: make(map[string]Page)}

func TestParsePage(t *testing.T) {
	file, _ := os.Open("testPage.html")
	seedUrl, _ := url.Parse("http://test.com/")
	page, _ := fetcher.ParsePage(file, seedUrl)

	if page.title != "TestPage Title" {
		t.Errorf("failed to parse page title, got %s, expected %s", page.title, "TestPage Title")
	}

	if len(page.urls) != 3 {
		t.Errorf("incorrect number of links found, got %d, expected %d", len(page.urls), 3)
	}

	if page.urls[0].String() != "http://test.com/page1" {
		t.Errorf("failed to parse first link, got %s, expected %s", page.urls[0].String(), "http://test.com/page1")
	}
}

func TestFetch(t *testing.T) {
	httpmock.Activate()

	file, _ := ioutil.ReadFile("testPage.html")
	mockresp := string(file)
	httpmock.RegisterResponder("GET", "http://test.com/",
		httpmock.NewStringResponder(200, mockresp))

	seedUrl, _ := url.Parse("http://test.com/")
	page, _ := fetcher.Fetch(seedUrl)

	if page.title != "TestPage Title" {
		t.Errorf("failed to parse page title, got %s, expected %s", page.title, "TestPage Title")
	}

	if len(fetcher.visited) != 1 {
		t.Errorf("incorrect link cache, got %d, expected %d", len(fetcher.visited), 1)
	}

	httpmock.DeactivateAndReset()
	page2, _ := fetcher.Fetch(seedUrl)

	if page2.title != "TestPage Title" {
		t.Errorf("faild to fetch page fromc cache, got %s, expected %s", page2.title, "TestPage Title")
	}

	if len(fetcher.visited) != 1 {
		t.Errorf("incorrect link cache, got %d, expected %d", len(fetcher.visited), 1)
	}
}

// I expect there is a  better way to test when using channels
func TestCrawl(t *testing.T) {
	result := make(chan string)
	var seedUrl, _ = url.Parse("http://test.com/")
	go Crawl(seedUrl, 0, 2, testFetcher, result)

	data := <-result
	if data != "|_ http://test.com/ \"Home\"" {
		t.Errorf("unexpectedResult %s", data)
	}
	data = <-result
	if data != "	|_ http://test.com/page1/ \"Page 1\"" {
		t.Errorf("unexpectedResult %s", data)
	}
	data = <-result
	if data != "		|_ http://test.com/page3/" {
		t.Errorf("unexpectedResult %s", data)
	}
	data = <-result
	if data != "	|_ http://test.com/page2/ \"Page 2\"" {
		t.Errorf("unexpectedResult %s", data)
	}
	data = <-result
	if data != "		|_ http://test.com/page4/" {
		t.Errorf("unexpectedResult %s", data)
	}

}

type fakeFetcher map[string]*Page

func (f fakeFetcher) Fetch(pageUrl *url.URL) (*Page, error) {
	return f[pageUrl.String()], nil
}
func fakeUrl(link string) *url.URL {
	val, _ := url.Parse(link)
	return val
}

var testFetcher = fakeFetcher{
	"http://test.com/": &Page{
		"Home",
		[]*url.URL{
			fakeUrl("http://test.com/page1/"),
			fakeUrl("http://test.com/page2/"),
		},
	},
	"http://test.com/page1/": &Page{
		"Page 1",
		[]*url.URL{
			fakeUrl("http://test.com/page3/"),
		},
	},
	"http://test.com/page2/": &Page{
		"Page 2",
		[]*url.URL{
			fakeUrl("http://test.com/page4/"),
		},
	},
}
