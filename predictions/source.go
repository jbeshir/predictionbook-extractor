package predictions

import (
	"context"
	"errors"
	"fmt"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"golang.org/x/net/html"
	"sort"
	"strconv"
	"strings"
	"time"
	"math"
	"github.com/golang/glog"
)

type Source struct {
	baseUrl   string
	extractor *htmlextract.Extractor
}

func NewSource(extractor *htmlextract.Extractor, baseUrl string) *Source {
	return &Source{
		baseUrl:   baseUrl,
		extractor: extractor,
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
	page, err := s.extractor.GetHtml(ctx, s.baseUrl+"/predictions/"+strconv.FormatInt(prediction, 10))
	if err != nil {
		return nil, err
	}

	responseNodes := htmlextract.HtmlNodesByAttr(page, "", "", "class", "response")
	for _, node := range responseNodes {

		response := new(PredictionResponse)
		response.Prediction = prediction

		timeNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "date")
		if timeNode != nil {
			for _, attr := range timeNode.Attr {
				if attr.Key == "title" {
					createdAt, err := time.Parse("2006-01-02 15:04:05 MST", attr.Val)
					if err == nil {
						response.Time = createdAt
					}
					break
				}
			}
		}

		userNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "user")
		if userNode != nil && userNode.FirstChild != nil && userNode.FirstChild.Type == html.TextNode {
			response.User = userNode.FirstChild.Data
		}

		confidenceNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "confidence")
		if confidenceNode != nil && confidenceNode.FirstChild != nil && confidenceNode.FirstChild.Type == html.TextNode {
			confidenceText := strings.TrimSpace(confidenceNode.FirstChild.Data)
			var confidencePercentage float64
			_, err := fmt.Sscanf(confidenceText, "%f%%", &confidencePercentage)
			if err == nil {
				response.Confidence = confidencePercentage / 100
			} else {
				response.Confidence = math.NaN()
			}
		} else {
			response.Confidence = math.NaN()
		}

		commentNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "comment")
		if commentNode != nil && commentNode.FirstChild != nil && commentNode.FirstChild.Type == html.TextNode {
			response.Comment = commentNode.FirstChild.Data
		}

		responses = append(responses, response)
	}
	return responses, nil
}

func (s *Source) RetrievePredictionListPage(ctx context.Context, index int64) (predictions []*PredictionSummary, pageInfo *PredictionPageInfo, err error) {
	page, err := s.extractor.GetHtml(ctx, s.baseUrl+"/predictions/page/"+strconv.FormatInt(index, 10))
	if err != nil {
		return nil, nil, err
	}

	predictionNodes := htmlextract.HtmlNodesByAttr(page, "", "", "class", "prediction")
	for _, node := range predictionNodes {
		prediction := new(PredictionSummary)

		titleNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "title")
		titleLinkNode := htmlextract.HtmlNodeByAttr(titleNode, "a", "", "", "")
		if titleLinkNode != nil {
			if titleLinkNode.FirstChild != nil && titleLinkNode.FirstChild.Type == html.TextNode {
				prediction.Title = titleLinkNode.FirstChild.Data
			}
			for _, attr := range titleLinkNode.Attr {
				if attr.Key == "href" {
					predictionUrl := attr.Val
					predictionUrlParts := strings.Split(predictionUrl, "/")
					if len(predictionUrlParts) > 0 {
						predictionIdStr := predictionUrlParts[len(predictionUrlParts)-1]
						id, err := strconv.ParseInt(predictionIdStr, 10, 64)
						if err == nil {
							prediction.Id = id
						}
					}
					break
				}
			}
		}

		creatorNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "creator")
		if creatorNode != nil && creatorNode.FirstChild != nil && creatorNode.FirstChild.Type == html.TextNode {
			prediction.Creator = creatorNode.FirstChild.Data
		}

		createdAtNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "created_at")
		if createdAtNode != nil {
			for _, attr := range createdAtNode.Attr {
				if attr.Key == "title" {
					createdAt, err := time.Parse("2006-01-02 15:04:05 MST", attr.Val)
					if err == nil {
						prediction.Created = createdAt
					}
					break
				}
			}
		}

		deadlineNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "deadline")
		deadlineDateNode := htmlextract.HtmlNodeByAttr(deadlineNode, "", "", "class", "date")
		if deadlineDateNode != nil {
			for _, attr := range deadlineDateNode.Attr {
				if attr.Key == "title" {
					deadline, err := time.Parse("2006-01-02 15:04:05 MST", attr.Val)
					if err == nil {
						prediction.Deadline = deadline
					}
					break
				}
			}
		}

		confidenceNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "mean_confidence")
		if confidenceNode != nil && confidenceNode.FirstChild != nil && confidenceNode.FirstChild.Type == html.TextNode {
			confidenceText := strings.TrimSpace(confidenceNode.FirstChild.Data)
			var confidencePercentage float64
			_, err := fmt.Sscanf(confidenceText, "%f%% confidence", &confidencePercentage)
			if err == nil {
				prediction.MeanConfidence = confidencePercentage / 100
			}
		}

		wagerCountNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "wagers_count")
		if wagerCountNode != nil && wagerCountNode.FirstChild != nil && wagerCountNode.FirstChild.Type == html.TextNode {
			wagerCountText := strings.TrimSpace(wagerCountNode.FirstChild.Data)
			var wagerCount int64
			_, err := fmt.Sscanf(wagerCountText, "%d wagers", &wagerCount)
			if err == nil {
				prediction.WagerCount = wagerCount
			} else {
				prediction.WagerCount = 1
			}
		} else {
			prediction.WagerCount = 1
		}

		outcomeNode := htmlextract.HtmlNodeByAttr(node, "", "", "class", "outcome")
		if outcomeNode != nil && outcomeNode.FirstChild != nil && outcomeNode.FirstChild.Type == html.TextNode {
			outcomeStr := strings.TrimSpace(outcomeNode.FirstChild.Data)
			if outcomeStr == "right" {
				prediction.Outcome = Right
			} else if outcomeStr == "wrong" {
				prediction.Outcome = Wrong
			} else {
				prediction.Outcome = Unknown
			}
		} else {
			prediction.Outcome = Unknown
		}

		predictions = append(predictions, prediction)
	}

	pageInfo = new(PredictionPageInfo)
	pageInfo.Index = index

	paginationNode := htmlextract.HtmlNodeByAttr(page, "nav", "", "class", "pagination")
	lastPageNode := htmlextract.HtmlNodeByAttr(paginationNode, "", "", "class", "last")
	linkNode := htmlextract.HtmlNodeByAttr(lastPageNode, "a", "", "", "")
	if linkNode != nil {
		for _, attr := range linkNode.Attr {
			if attr.Key == "href" {
				if strings.HasPrefix(attr.Val, "/predictions/page/") {
					lastPageStr := attr.Val[len("/predictions/page/"):]
					lastPage, err := strconv.ParseInt(lastPageStr, 10, 64)
					if err == nil {
						pageInfo.LastPage = lastPage
					}
				}
			}
		}
	} else if lastPageNode == nil {
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
