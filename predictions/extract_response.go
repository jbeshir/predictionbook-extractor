package predictions

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"math"
	"strings"
	"time"
)

type PredictionResponse struct {
	Prediction int64
	Time       time.Time
	User       string
	Confidence float64
	Comment    string
}

func ExtractPredictionResponse(responseNode *html.Node, prediction int64) (response *PredictionResponse) {
	response = new(PredictionResponse)
	response.Prediction = prediction

	responseSelector := goquery.NewDocumentFromNode(responseNode).Selection

	response.Time = extractResponseTime(responseSelector)
	response.User = responseSelector.Find(".user").Text()
	response.Comment = responseSelector.Find(".comment").Text()
	response.Confidence = extractResponseConfidence(responseSelector)

	return
}

func extractResponseTime(responseSelector *goquery.Selection) (t time.Time) {
	createdAtStr, exists := responseSelector.Find(".date").Attr("title")
	if exists {
		createdAt, err := time.Parse("2006-01-02 15:04:05 MST", createdAtStr)
		if err == nil {
			t = createdAt
		}
	}

	return
}

func extractResponseConfidence(responseSelector *goquery.Selection) (confidence float64) {
	confidenceStr := strings.TrimSpace(responseSelector.Find(".confidence").Text())
	var confidencePercentage float64
	_, err := fmt.Sscanf(confidenceStr, "%f%%", &confidencePercentage)
	if err == nil {
		confidence = confidencePercentage / 100
	} else {
		confidence = math.NaN()
	}

	return
}
