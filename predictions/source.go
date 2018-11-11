package predictions

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
	"golang.org/x/net/html"
	"math"
	"sort"
	"strconv"
	"strings"
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

	page.Find(".response").Each(func(i int, responseNode *goquery.Selection) {

		response := new(PredictionResponse)
		response.Prediction = prediction

		createdAtStr, exists := responseNode.Find(".date").Attr("title")
		if exists {
			createdAt, err := time.Parse("2006-01-02 15:04:05 MST", createdAtStr)
			if err == nil {
				response.Time = createdAt
			}
		}

		response.User = responseNode.Find(".user").Text()
		response.Comment = responseNode.Find(".comment").Text()

		confidenceStr := strings.TrimSpace(responseNode.Find(".confidence").Text())
		var confidencePercentage float64
		_, err := fmt.Sscanf(confidenceStr, "%f%%", &confidencePercentage)
		if err == nil {
			response.Confidence = confidencePercentage / 100
		} else {
			response.Confidence = math.NaN()
		}

		responses = append(responses, response)
	})

	return responses, nil
}

func (s *Source) RetrievePredictionListPage(ctx context.Context, index int64) (predictions []*PredictionSummary, pageInfo *PredictionPageInfo, err error) {
	rootNode, err := s.htmlFetcher.GetHtml(ctx, s.baseUrl+"/predictions/page/"+strconv.FormatInt(index, 10))
	if err != nil {
		return nil, nil, err
	}
	page := goquery.NewDocumentFromNode(rootNode)

	page.Find(".prediction").Each(func(i int, predictionNode *goquery.Selection) {
		prediction := new(PredictionSummary)

		titleLink := predictionNode.Find(".title a")
		prediction.Title = titleLink.Text()
		predictionUrl, exists := titleLink.Attr("href")
		if exists {
			predictionUrlParts := strings.Split(predictionUrl, "/")
			if len(predictionUrlParts) > 0 {
				predictionIdStr := predictionUrlParts[len(predictionUrlParts)-1]
				id, err := strconv.ParseInt(predictionIdStr, 10, 64)
				if err == nil {
					prediction.Id = id
				}
			}
		}

		creator := predictionNode.Find(".creator")
		if len(creator.Nodes) > 0 {
			creatorNode := creator.Nodes[0]
			if creatorNode.FirstChild != nil && creatorNode.FirstChild.Type == html.TextNode {
				prediction.Creator = creatorNode.FirstChild.Data
			}
		}

		createdAtStr, exists := predictionNode.Find(".created_at").Attr("title")
		if exists {
			createdAt, err := time.Parse("2006-01-02 15:04:05 MST", createdAtStr)
			if err == nil {
				prediction.Created = createdAt
			}
		}

		deadlineStr, exists := predictionNode.Find(".deadline .date").Attr("title")
		if exists {
			deadline, err := time.Parse("2006-01-02 15:04:05 MST", deadlineStr)
			if err == nil {
				prediction.Deadline = deadline
			}
		}

		confidenceStr := strings.TrimSpace(predictionNode.Find(".mean_confidence").Text())
		var confidencePercentage float64
		_, err := fmt.Sscanf(confidenceStr, "%f%% confidence", &confidencePercentage)
		if err == nil {
			prediction.MeanConfidence = confidencePercentage / 100
		}

		wagerCountStr := strings.TrimSpace(predictionNode.Find(".wagers_count").Text())
		var wagerCount int64
		_, err = fmt.Sscanf(wagerCountStr, "%d wagers", &wagerCount)
		if err == nil {
			prediction.WagerCount = wagerCount
		} else {
			prediction.WagerCount = 1
		}

		outcomeStr := strings.TrimSpace(predictionNode.Find(".outcome").Text())
		if outcomeStr == "right" {
			prediction.Outcome = Right
		} else if outcomeStr == "wrong" {
			prediction.Outcome = Wrong
		} else {
			prediction.Outcome = Unknown
		}

		predictions = append(predictions, prediction)
	})

	pageInfo = new(PredictionPageInfo)
	pageInfo.Index = index

	lastPageHref, exists := page.Find("nav.pagination .last a").Attr("href")
	if exists {
		if strings.HasPrefix(lastPageHref, "/predictions/page/") {
			lastPageStr := lastPageHref[len("/predictions/page/"):]
			lastPage, err := strconv.ParseInt(lastPageStr, 10, 64)
			if err == nil {
				pageInfo.LastPage = lastPage
			}
		}
	} else {
		pageInfo.LastPage = index
	}

	return predictions, pageInfo, nil
}

type PredictionSummary struct {
	Id             int64
	Title          string
	Creator        string
	Created        time.Time
	Deadline       time.Time
	MeanConfidence float64
	WagerCount     int64
	Outcome        Outcome
}

type PredictionPageInfo struct {
	Index    int64
	LastPage int64
}

type PredictionResponse struct {
	Prediction int64
	Time       time.Time
	User       string
	Confidence float64
	Comment    string
}

type Outcome int64

const (
	Unknown Outcome = iota
	Right
	Wrong
)
