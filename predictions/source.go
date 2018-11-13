package predictions

import (
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"golang.org/x/net/html"
	"sort"
	"strconv"
)

type Source struct {
	baseUrl     string
	htmlFetcher HtmlFetcher
}

type HtmlFetcher interface {
	GetHtml(ctx context.Context, url string) (*html.Node, error)
}

func NewSource(htmlFetcher HtmlFetcher, baseUrl string) *Source {
	return &Source{
		baseUrl:     baseUrl,
		htmlFetcher: htmlFetcher,
	}
}

func (s *Source) Latest(ctx context.Context) (*PredictionSummary, error) {
	latest, _, err := s.RetrievePredictionListPage(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(latest) == 0 {
		return nil, errors.New("no predictions found")
	}

	return latest[0], nil
}

func (s *Source) PredictionPageCount(ctx context.Context) (int64, error) {
	_, info, err := s.RetrievePredictionListPage(ctx, 1)
	if err != nil {
		return 0, err
	}
	if info.LastPage == 0 {
		return 0, errors.New("unable to extract page count")
	}

	return info.LastPage, nil
}

func (s *Source) AllPredictions(ctx context.Context) (predictions []*PredictionSummary, err error) {
	currentPage := int64(1)
	totalPages := int64(1)
	for {
		newPredictions, pageInfo, err := s.RetrievePredictionListPage(ctx, currentPage)
		if err != nil {
			return nil, err
		}

		predictions = append(predictions, newPredictions...)
		totalPages = pageInfo.LastPage

		if currentPage == totalPages {
			break
		}
		currentPage++
	}

	// Sort and remove duplicates; they can appear due to predictions made during retrieval
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Id < predictions[j].Id
	})
	for i := 1; i < len(predictions); i++ {
		if predictions[i].Id == predictions[i-1].Id {
			predictions = append(predictions[:i], predictions[i+1:]...)
			i--
		}
	}

	return
}

func (s *Source) AllPredictionResponses(ctx context.Context, predictions []*PredictionSummary) (responses []*PredictionResponse, err error) {
	respCh := make(chan []*PredictionResponse, 1)
	errCh := make(chan error)

	launched := 0
	for _, p := range predictions {
		go func(prediction int64) {
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				var r []*PredictionResponse
				r, err = s.RetrievePredictionResponses(ctx, prediction)
				if err == nil {
					respCh <- r
					break
				}
			}
			if err != nil {
				errCh <- err
				return
			}
		}(p.Id)
		launched++
	}

	for i := 0; i < launched; i++ {
		select {
		case err := <-errCh:
			if glog.V(2) {
				glog.Infof("Got an error while generating responses: %s\n", err)
			}
			return nil, err
		case r := <-respCh:
			responses = append(responses, r...)
			if glog.V(2) {
				glog.Infof("Added %d responses, now collected %d responses...\n", len(r), len(responses))
			}
		}
	}

	if glog.V(2) {
		glog.Infof("Finished collecting responses\n")
	}
	return responses, nil
}

func (s *Source) RetrievePredictionResponses(ctx context.Context, prediction int64) (responses []*PredictionResponse, err error) {
	rootNode, err := s.htmlFetcher.GetHtml(ctx, s.baseUrl+"/predictions/"+strconv.FormatInt(prediction, 10))
	if err != nil {
		return nil, err
	}

	page := goquery.NewDocumentFromNode(rootNode)
	page.Find(".response").Each(func(i int, responseSelector *goquery.Selection) {
		response := ExtractPredictionResponse(responseSelector.Nodes[0], prediction)
		responses = append(responses, response)
	})

	return responses, nil
}

func (s *Source) RetrievePredictionListPage(ctx context.Context, index int64) (predictions []*PredictionSummary, pageInfo *PredictionListPageInfo, err error) {
	rootNode, err := s.htmlFetcher.GetHtml(ctx, s.baseUrl+"/predictions/page/"+strconv.FormatInt(index, 10))
	if err != nil {
		return nil, nil, err
	}
	page := goquery.NewDocumentFromNode(rootNode)

	page.Find(".prediction").Each(func(i int, predictionSelector *goquery.Selection) {
		prediction := ExtractPredictionSummary(predictionSelector.Nodes[0])
		predictions = append(predictions, prediction)
	})

	pageInfo = ExtractPredictionListPageInfo(rootNode, index)

	return predictions, pageInfo, nil
}
