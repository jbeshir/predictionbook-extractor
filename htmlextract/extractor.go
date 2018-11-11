package htmlextract

import (
	"context"
	"errors"
	"github.com/golang/glog"
	"golang.org/x/net/html"
	"golang.org/x/time/rate"
	"net/http"
)

type Extractor struct {
	limiter          *rate.Limiter
	requestTokenPool chan bool
}

func NewExtractor(limiter *rate.Limiter, concurrentRequestLimit int) *Extractor {
	e := &Extractor{
		limiter:          limiter,
		requestTokenPool: make(chan bool, concurrentRequestLimit),
	}

	for i := 0; i < concurrentRequestLimit; i++ {
		e.requestTokenPool <- true
	}

	return e
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
		if glog.V(2) {
			glog.Infof("HTTP request error: %s\n", err)
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if glog.V(2) {
			glog.Infof("HTTP error: %s\n", resp.Status)
		}
		return nil, errors.New("HTTP error: " + resp.Status)
	}

	rootNode, err := html.Parse(resp.Body)
	if err != nil {
		if glog.V(2) {
			glog.Infof("Parse error: %s\n", err.Error())
		}
		return nil, errors.New("Parse error: " + err.Error())
	}

	if glog.V(2) {
		glog.Infoln("Retrieved", url)
	}
	return rootNode, nil
}
