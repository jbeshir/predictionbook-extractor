package predictions

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"math"
	"strings"
	"time"
)

func ExtractPredictionSummaryResponsePage(id int64, responsePageNode *html.Node) (prediction *PredictionSummary) {
	responsePage := goquery.NewDocumentFromNode(responsePageNode).Selection

	prediction = new(PredictionSummary)
	prediction.Id = id
	prediction.Title = strings.TrimSpace(responsePage.Find("h1").Text())
	prediction.WagerCount = int64(len(responsePage.Find(".response").Nodes))
	prediction.Creator = responsePage.Find("#content > p > a.user").Text()
	prediction.Created = extractSummaryResponsePageCreated(responsePage)
	prediction.Deadline = extractSummaryResponsePageDeadline(responsePage)
	prediction.Outcome = extractSummaryOutcome(responsePage)

	sumConfidence := 0.0
	totalAssignments := 0
	for _, respNode := range responsePage.Find(".response").Nodes {
		resp := ExtractPredictionResponse(respNode, id)
		if !math.IsNaN(resp.Confidence) {
			sumConfidence += resp.Confidence
			totalAssignments++
		}
	}
	prediction.MeanConfidence = sumConfidence / float64(totalAssignments)
	return
}

func extractSummaryResponsePageCreated(pageSelector *goquery.Selection) (created time.Time) {
	createdAtStr, exists := pageSelector.Find("#content > p > .date").First().Attr("title")
	if exists {
		t, err := time.Parse("2006-01-02 15:04:05 MST", createdAtStr)
		if err == nil {
			created = t
		}
	}

	return
}

func extractSummaryResponsePageDeadline(pageSelector *goquery.Selection) (deadline time.Time) {
	deadlineStr, exists := pageSelector.Find("#content > p > .date").Last().Attr("title")
	if exists {
		t, err := time.Parse("2006-01-02 15:04:05 MST", deadlineStr)
		if err == nil {
			deadline = t
		}
	}

	return
}