package predictions

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"strconv"
	"strings"
	"time"
)

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

type Outcome int64

const (
	Unknown Outcome = iota
	Right
	Wrong
)

func ExtractPredictionSummary(predictionNode *html.Node) (prediction *PredictionSummary) {
	prediction = new(PredictionSummary)

	predictionSelector := goquery.NewDocumentFromNode(predictionNode).Selection

	prediction.Title, prediction.Id = extractSummaryTitleAndId(predictionSelector)
	prediction.Creator = extractSummaryCreator(predictionSelector)
	prediction.Created = extractSummaryCreated(predictionSelector)
	prediction.Deadline = extractSummaryDeadline(predictionSelector)
	prediction.MeanConfidence = extractSummaryMeanConfidence(predictionSelector)
	prediction.WagerCount = extractSummaryWagerCount(predictionSelector)
	prediction.Outcome = extractSummaryOutcome(predictionSelector)

	return
}

func extractSummaryTitleAndId(predictionSelector *goquery.Selection) (title string, id int64) {

	titleLink := predictionSelector.Find(".title a")
	title = titleLink.Text()

	predictionUrl, exists := titleLink.Attr("href")
	if exists {
		predictionUrlParts := strings.Split(predictionUrl, "/")
		if len(predictionUrlParts) > 0 {
			predictionIdStr := predictionUrlParts[len(predictionUrlParts)-1]
			parsedId, err := strconv.ParseInt(predictionIdStr, 10, 64)
			if err == nil {
				id = parsedId
			}
		}
	}

	return
}

func extractSummaryCreator(predictionSelector *goquery.Selection) (creator string) {
	creatorSelector := predictionSelector.Find(".creator")
	if len(creatorSelector.Nodes) > 0 {
		creatorNode := creatorSelector.Nodes[0]
		if creatorNode.FirstChild != nil && creatorNode.FirstChild.Type == html.TextNode {
			creator = creatorNode.FirstChild.Data
		}
	}

	return
}

func extractSummaryCreated(predictionSelector *goquery.Selection) (created time.Time) {
	createdAtStr, exists := predictionSelector.Find(".created_at").Attr("title")
	if exists {
		t, err := time.Parse("2006-01-02 15:04:05 MST", createdAtStr)
		if err == nil {
			created = t
		}
	}

	return
}

func extractSummaryDeadline(predictionSelector *goquery.Selection) (deadline time.Time) {
	deadlineStr, exists := predictionSelector.Find(".deadline .date").Attr("title")
	if exists {
		t, err := time.Parse("2006-01-02 15:04:05 MST", deadlineStr)
		if err == nil {
			deadline = t
		}
	}

	return
}

func extractSummaryMeanConfidence(predictionSelector *goquery.Selection) (meanConfidence float64) {
	confidenceStr := strings.TrimSpace(predictionSelector.Find(".mean_confidence").Text())
	var confidencePercentage float64
	_, err := fmt.Sscanf(confidenceStr, "%f%% confidence", &confidencePercentage)
	if err == nil {
		meanConfidence = confidencePercentage / 100
	}

	return
}

func extractSummaryWagerCount(predictionSelector *goquery.Selection) (wagerCount int64) {
	wagerCountStr := strings.TrimSpace(predictionSelector.Find(".wagers_count").Text())
	var i int64
	_, err := fmt.Sscanf(wagerCountStr, "%d wagers", &i)
	if err == nil {
		wagerCount = i
	} else {
		wagerCount = 1
	}

	return
}

func extractSummaryOutcome(predictionSelector *goquery.Selection) (outcome Outcome) {
	outcomeStr := strings.TrimSpace(predictionSelector.Find(".outcome").First().Text())
	if outcomeStr == "right" {
		outcome = Right
	} else if outcomeStr == "wrong" {
		outcome = Wrong
	} else {
		outcome = Unknown
	}

	return
}
