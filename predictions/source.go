package predictions

import (
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"golang.org/x/net/html"
	"sort"
	"strconv"
	"time"
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

	return info.LastPage, nil
}

func (s *Source) AllPredictions(ctx context.Context) (predictions []*PredictionSummary, err error) {

	return s.AllPredictionsSince(ctx, time.Time{})
}

func (s *Source) AllPredictionsSince(ctx context.Context, t time.Time) (predictions []*PredictionSummary, err error) {

	currentPage := int64(1)
	totalPages := int64(1)
	for {
		newPredictions, pageInfo, err := s.RetrievePredictionListPage(ctx, currentPage)
		if err != nil {
			return nil, err
		}

		lastIncluded := len(newPredictions) - 1
		for lastIncluded > -1 && newPredictions[lastIncluded].Created.Before(t) {
			lastIncluded--
		}

		predictions = append(predictions, newPredictions[:lastIncluded+1]...)

		if lastIncluded < len(newPredictions)-1 {
			break
		}

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

func (s *Source) AllPredictionResponses(ctx context.Context, predictions []*PredictionSummary) (summaries []*PredictionSummary, responses []*PredictionResponse, err error) {
	respCh := make(chan struct{s *PredictionSummary; rs []*PredictionResponse}, 1)
	errCh := make(chan error)

	launched := 0
	for _, p := range predictions {
		go func(prediction int64) {
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				var rs []*PredictionResponse
				var sum *PredictionSummary
				sum, rs, err = s.RetrievePredictionResponses(ctx, prediction)
				if err == nil {
					respCh <- struct{s *PredictionSummary; rs []*PredictionResponse}{s:sum,rs:rs}
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
			return nil, nil, err
		case r := <-respCh:
			summaries = append(summaries, r.s)
			responses = append(responses, r.rs...)
			if glog.V(2) {
				glog.Infof("Added %d responses, now collected %d responses...\n", len(r.rs), len(responses))
			}
		}
	}

	if glog.V(2) {
		glog.Infof("Finished collecting responses\n")
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Id < summaries[j].Id
	})
	sort.Slice(responses, func(i, j int) bool {
		if responses[i].Prediction != responses[j].Prediction {
			return responses[i].Prediction < responses[j].Prediction
		}
		return responses[i].Time.Unix() < responses[j].Time.Unix()
	})

	return summaries, responses, nil
}

func (s *Source) RetrievePredictionResponses(ctx context.Context, prediction int64) (summary *PredictionSummary, responses []*PredictionResponse, err error) {
	rootNode, err := s.htmlFetcher.GetHtml(ctx, s.baseUrl+"/predictions/"+strconv.FormatInt(prediction, 10))
	if err != nil {
		return nil, nil, err
	}

	page := goquery.NewDocumentFromNode(rootNode)
	page.Find(".response").Each(func(i int, responseSelector *goquery.Selection) {
		response := ExtractPredictionResponse(responseSelector.Nodes[0], prediction)
		responses = append(responses, response)
	})

	return ExtractPredictionSummaryResponsePage(prediction, rootNode), responses, nil
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
