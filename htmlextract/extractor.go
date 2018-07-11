package htmlextract

import (
	"golang.org/x/net/html"
	"strings"
	"github.com/golang/glog"
	"net/http"
	"errors"
	"golang.org/x/time/rate"
	"context"
)

type Extractor struct {
	limiter *rate.Limiter
	requestTokenPool chan bool
}

func NewExtractor(limiter *rate.Limiter, concurrentRequestLimit int) *Extractor {
	e := &Extractor{
		limiter: limiter,
		requestTokenPool: make(chan bool, concurrentRequestLimit),
	}

	for i := 0; i < concurrentRequestLimit; i++ {
		e.requestTokenPool <- true
	}

	return e
}

func (e *Extractor) HtmlNodesByAttrInPages(ctx context.Context, urls []string, tag, namespace, attr, val string) (results []*html.Node, err error) {

	// Create channel for receiving per-page results.
	resultCh := make(chan struct {
		Results []*html.Node
		Err     error
	}, len(urls))

	// Search pages in parallel.
	for _, url := range urls {
		go func(url string) {

			var result struct {
				Results []*html.Node
				Err     error
			}
			if err != nil {
				result.Err = err
				resultCh <- result
				return
			}

			rootNode, err := e.GetHtml(ctx, url)
			if err != nil {
				result.Err = err
				resultCh <- result
				return
			}

			result.Results = HtmlNodesByAttr(rootNode, namespace, tag, attr, val)
			resultCh <- result
		}(url)
	}

	// Collect and combine per-page results,
	// return when we got them all or on error.
	resultCount := 0
	for result := range resultCh {

		if result.Err != nil {
			return nil, errors.New(result.Err.Error())
		}

		results = append(results, result.Results...)

		resultCount++
		if resultCount >= len(urls) {
			break
		}
	}

	return
}

func HtmlNodesByAttr(parentNode *html.Node, tag, namespace, key, val string) (results []*html.Node) {

	var stack []*html.Node
	currentNode := parentNode

	for currentNode != nil {

		if tag == "" || (currentNode.Type == html.ElementNode && currentNode.Data == tag) {

			if key != "" {
				// Check if our current node matches.
				for _, attr := range currentNode.Attr {
					if attr.Namespace != namespace {
						continue
					}
					if attr.Key != key {
						continue
					}

					if attr.Val == val {
						results = append(results, currentNode)
						break
					}

					partsVal := strings.Split(attr.Val, " ")
					for _, partVal := range partsVal {
						if partVal == val {
							results = append(results, currentNode)
							break
						}
					}
					break
				}
			} else {
				results = append(results, currentNode)
			}
		}

		// If our current node has children,
		// push it onto the stack and go to first child.
		// Otherwise, proceed to the next sibling.
		if currentNode.FirstChild != nil {
			stack = append(stack, currentNode)
			currentNode = currentNode.FirstChild
		} else {
			// If we've reached the end of our list of siblings,
			// and we can go up stack, do so.
			for currentNode.NextSibling == nil && len(stack) > 0 {
				currentNode = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}

			// Advance to our next sibling.
			// This will set the current node to nil,
			// at the end of the document.
			currentNode = currentNode.NextSibling
		}
	}

	return
}

func (e *Extractor) GetHtml(ctx context.Context, url string) (*html.Node, error) {

	err := e.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	<-e.requestTokenPool
	defer func() {
		e.requestTokenPool <- true
	}()

	if glog.V(2) {
		glog.Infoln("Retrieving", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("HTTP error: " + resp.Status)
	}

	rootNode, err := html.Parse(resp.Body)
	if err != nil {
		return nil, errors.New("Parse error: " + err.Error())
	}

	return rootNode, nil
}