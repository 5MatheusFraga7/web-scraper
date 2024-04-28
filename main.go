package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Result struct {
	Title string
}

func (r Result) String() string {
	return fmt.Sprint(r.Title)
}

func hasClass(attribs []html.Attribute, className string) bool {
	for _, attr := range attribs {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			return true
		}
	}
	return false
}

func main() {
	urlToProcess := []string{
		"https://pt.wikipedia.org/wiki/A_Era_do_Apocalipse",
		"https://pt.wikipedia.org/wiki/Banshee_(Marvel_Comics)",
		"https://pt.wikipedia.org/wiki/Cable",
		"https://pt.wikipedia.org/wiki/Colossus_(Marvel_Comics)",
		"https://pt.wikipedia.org/wiki/Lockheed_(Marvel_Comics)",
		"https://pt.wikipedia.org/wiki/Rogue",
	}

	ini := time.Now()
	r := make(chan Result)
	go scrapListURL(urlToProcess, r)
	fmt.Println("With goroutines:")
	for url := range r {
		fmt.Println(url)
	}

	fmt.Println("(Took ", time.Since(ini).Seconds(), "secs)")
}

func scrapListURL(urlToProcess []string, rchan chan Result) {
	defer close(rchan)
	var results = []chan Result{}

	for i, url := range urlToProcess {
		results = append(results, make(chan Result))
		go scrapParallel(url, results[i])
	}

	for i := range results {
		for r1 := range results[i] {
			rchan <- r1
		}
	}
}

func scrapParallel(url string, rchan chan Result) {
	defer close(rchan)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("ERROR: It can't scrap '", url, "'")
	}
	defer resp.Body.Close()
	body := resp.Body
	htmlParsed, err := html.Parse(body)
	if err != nil {
		fmt.Println("ERROR: It can't parse html '", url, "'")
	}

	var r Result

	div := getFirstElementByClass(htmlParsed, "span", "mw-page-title-main")

	textNode := getFirstTextNode(div)

	if textNode != nil {
		r.Title = textNode.Data
	}

	rchan <- r
}

func getFirstElementByClass(htmlParsed *html.Node, elm, className string) *html.Node {

	if htmlParsed == nil {
		return nil
	}

	for m := htmlParsed.FirstChild; m != nil; m = m.NextSibling {
		if m.Data == elm && hasClass(m.Attr, className) {
			return m
		}
		r := getFirstElementByClass(m, elm, className)
		if r != nil {
			return r
		}
	}
	return nil
}

func getFirstTextNode(htmlParsed *html.Node) *html.Node {
	if htmlParsed == nil {
		return nil
	}

	for m := htmlParsed.FirstChild; m != nil; m = m.NextSibling {
		if m.Type == html.TextNode {
			return m
		}
		r := getFirstTextNode(m)
		if r != nil {
			return r
		}
	}
	return nil
}

func printHTML(node *html.Node, depth int) {
	if node == nil {
		return
	}

	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}
	fmt.Printf("<%s", node.Data)
	for _, attr := range node.Attr {
		fmt.Printf(" %s=\"%s\"", attr.Key, attr.Val)
	}
	fmt.Println(">")

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		printHTML(child, depth+1)
	}

	if node.Type == html.ElementNode {
		for i := 0; i < depth; i++ {
			fmt.Print("  ")
		}
		fmt.Printf("</%s>\n", node.Data)
	}
}
