package predictions

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"math"
	"testing"
)

func testSummariesLoad(t *testing.T) (summaryNodes []*html.Node) {
	rootNode := testHtmlLoad(t, "test_list.html")

	page := goquery.NewDocumentFromNode(rootNode)
	page.Find(".prediction").Each(func(i int, s *goquery.Selection) {
		summaryNodes = append(summaryNodes, s.Nodes[0])
	})
	if len(summaryNodes) != 50 {
		t.Fatalf("testSummariesLoad: Couldn't load prediction summaries, expected %d summaries, got %d", 50, len(summaryNodes))
	}

	return
}

func TestExtractPredictionSummarySingleWager(t *testing.T) {
	t.Parallel()

	summaries := testSummariesLoad(t)
	prediction := ExtractPredictionSummary(summaries[3])

	if prediction.Outcome != Wrong {
		t.Errorf("Incorrect prediction outcome, should be %d, was %d", Wrong, prediction.Outcome)
	}
	if prediction.Id != 193469 {
		t.Errorf("Incorrect prediction ID, should be %d, was %d", 193469, prediction.Id)
	}
	if prediction.Title != "I get at least a raise of at least 9%." {
		t.Errorf("Incorrect prediction title, should be %s, was %s", "I get at least a raise of at least 9%.", prediction.Title)
	}
	if prediction.Creator != "notsonewuser" {
		t.Errorf("Incorrect prediction creator, should be %s, was %s", "notsonewuser", prediction.Creator)
	}
	if prediction.Created.Unix() != 1539214517 {
		t.Errorf("Incorrect prediction created time, should be %d, was %d", 1539214517, prediction.Created.Unix())
	}
	if prediction.Deadline.Unix() != 1541088000 {
		t.Errorf("Incorrect prediction deadline, should be %d, was %d", 1541088000, prediction.Deadline.Unix())
	}
	if math.Abs(prediction.MeanConfidence-0.3) > 0.00001 {
		t.Errorf("Incorrect prediction mean confidence, should be %g, was %g", 0.3, prediction.MeanConfidence)
	}
	if prediction.WagerCount != 1 {
		t.Errorf("Incorrect prediction wager count, should be %d, was %d", 1, prediction.WagerCount)
	}
}

func TestExtractPredictionSummaryMultipleWager(t *testing.T) {
	t.Parallel()

	summaries := testSummariesLoad(t)
	prediction := ExtractPredictionSummary(summaries[6])
	if prediction.WagerCount != 7 {
		t.Errorf("Incorrect prediction wager count, should be %d, was %d", 7, prediction.WagerCount)
	}
}
