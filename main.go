package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-collections/collections/stack"
	"golang.org/x/net/html"
)

type UrlData struct {
	Url   string
	Depth int
}

func parseLinks(urlStack *stack.Stack, htmlDoc *html.Node, currentUrlData *UrlData) error {
	//fmt.Println(*currentUrl)
	// to process siblings
	for tag := htmlDoc.FirstChild; tag != nil; tag = tag.NextSibling {
		// checking if the tag is an elment and a tag
		if tag.Type == html.ElementNode && tag.Data == "a" {
			for _, a := range tag.Attr {
				if a.Key == "href" {
					finalUrl, valid := processUrl(currentUrlData.Url, a.Val)
					if !valid {
						fmt.Println("Invalid", finalUrl)
						continue
					}
					// will be empty in case url was "/" that is root or baseUrl
					if finalUrl != "" {
						var newUrlData UrlData
						newUrlData.Depth = (*currentUrlData).Depth + 1
						newUrlData.Url = finalUrl
						urlStack.Push(newUrlData)
					} else {
						urlStack.Push(parseBaseUrl((*currentUrlData).Url))
					}
				}
			}
			//} else if tag.Type == html.ElementNode && tag.Data == "body" {
			//fmt.Println(renderNode(tag))
		}
		// to process childs at the same level
		if tag.FirstChild != nil {
			parseLinks(urlStack, tag, currentUrlData)
		}
	}
	return nil
}

func processUrl(currentUrl string, link string) (string, bool) {
	var finalUrl string
	var valid bool
	if strings.Index(link, "./") == 0 {
		// url is relative and in current path, append baseUrl to it
		finalUrl = currentUrl + "/" + strings.TrimLeft(link, "./")
		valid = true
	} else if strings.Index(link, "/") == 0 {
		// url is relative and exists at root path, baseUrl will be required
		baseUrl := parseBaseUrl(currentUrl)
		finalUrl = baseUrl + "/" + strings.TrimLeft(link, "/")
		valid = true
	} else if strings.Index(link, "../") == 0 {
		// url is relative and exists at one path behind the current one
		// need to trim the one path back two cases are possible '/foo/bar' and '/foo/bar/' the result shall convert to '/foo'
		currentUrl = strings.TrimRight(currentUrl, "/")
		currentUrl = currentUrl[0:strings.LastIndex(currentUrl, "/")]
		finalUrl = currentUrl + "/" + strings.TrimLeft(link, "../")
		valid = true
	} else {
		finalUrl = link
	}
	valid = strings.Contains(strings.ToLower(finalUrl), "http")
	return finalUrl, valid
}

func parseBaseUrl(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err)
	}
	u.Path = ""
	u.Fragment = ""
	u.RawQuery = ""
	return u.String()
}

// CrawlWebpage craws the given rootURL looking for <a href=""> tags
// that are targeting the current web page, either via an absolute url like http://mysite.com/mypath or by a relative url like /mypath
// and returns a sorted list of absolute urls  (eg: []string{"http://mysite.com/1","http://mysite.com/2"})
func CrawlWebpage(rootURL string, maxDepth int) ([]string, error) {
	//TODO: Implement Solution
	//fmt.Println(rootURL, maxDepth)
	urlData := UrlData{
		Url:   rootURL,
		Depth: 0,
	}
	var urlStack stack.Stack
	if rootURL != "" {
		urlStack.Push(urlData)
	}

	var finalUrlArr []string

	for urlStack.Len() > 0 {
		link := urlStack.Pop()

		linkData := link.(UrlData)
		fmt.Println(linkData)
		if linkData.Depth >= maxDepth {
			continue
		}
		finalUrlArr = append(finalUrlArr, linkData.Url)
		resp, err := http.Get(linkData.Url)
		if err != nil {
			log.Println("Failed:", err)
			continue
		}
		defer resp.Body.Close()

		htmlDoc, err := html.Parse(resp.Body)
		if err != nil {
			log.Println("Failed:", err)
		}
		err = parseLinks(&urlStack, htmlDoc, &linkData)
		if err != nil {
			log.Println("Failed:", err)
		}
	}
	return finalUrlArr, nil
}

// --- DO NOT MODIFY BELOW ---

func main() {
	const (
		defaultURL      = "https://www.example.com/"
		defaultMaxDepth = 3
	)
	urlFlag := flag.String("url", defaultURL, "the url that you want to crawl")
	maxDepth := flag.Int("depth", defaultMaxDepth, "the maximum number of links deep to traverse")
	flag.Parse()

	links, err := CrawlWebpage(*urlFlag, *maxDepth)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	fmt.Println("Links")
	fmt.Println("-----")
	for i, l := range links {
		fmt.Printf("%03d. %s\n", i+1, l)
	}
	fmt.Println()
}
