package predictions

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"math"
	"testing"
)

func testResponsesLoad(t *testing.T) (responseNodes []*html.Node) {
	rootNode := testHtmlLoad(t, "test_responses.html")

	page := goquery.NewDocumentFromNode(rootNode)
	page.Find(".response").Each(func(i int, s *goquery.Selection) {
		responseNodes = append(responseNodes, s.Nodes[0])
	})
	if len(responseNodes) != 8 {
		t.Fatalf("testSummariesLoad: Couldn't load prediction responses, expected %d responses, got %d", 8, len(responseNodes))
	}

	return
}

func TestExtractResponseAssignmentOnly(t *testing.T) {
	responses := testResponsesLoad(t)
	response := ExtractPredictionResponse(responses[1], 193436)

	if response.Prediction != 193436 {
		t.Errorf("Incorrect response prediction ID, should be %d, was %d", 193436, response.Prediction)
	}
	if response.User != "pranomostro" {
		t.Errorf("Incorrect response user, should be %s, was %s", "pranomostro", response.User)
	}
	if response.Time.Unix() != 1539248198 {
		t.Errorf("Incorrect response time, should be %d, was %d", 1539248198, response.Time.Unix())
	}
	if math.Abs(response.Confidence-0.25) > 0.00001 {
		t.Errorf("Incorrect response confidence, should be %g, was %g", 0.25, response.Confidence)
	}
	if response.Comment != "" {
		t.Errorf("Incorrect response comment, should be empty string, was %s", response.Comment)
	}
}

func TestExtractResponseCommentOnly(t *testing.T) {
	responses := testResponsesLoad(t)
	response := ExtractPredictionResponse(responses[2], 193436)

	if !math.IsNaN(response.Confidence) {
		t.Errorf("Incorrect response confidence, should be NaN, was %g", response.Confidence)
	}
	if response.Comment != "Compromised in general, or specifically compromised in the way described by the recent Bloomberg article?" {
		t.Errorf("Incorrect response comment, should be '%s', was '%s'", "Compromised in general, or specifically compromised in the way described by the recent Bloomberg article?", response.Comment)
	}
}

func TestExtractResponseCommentAndAssignment(t *testing.T) {
	responses := testResponsesLoad(t)
	response := ExtractPredictionResponse(responses[5], 193436)

	if math.Abs(response.Confidence-0.15) > 0.00001 {
		t.Errorf("Incorrect response confidence, should be %g, was %g", 0.15, response.Confidence)
	}
	if response.Comment != "I’m assuming you mean specifically compromised in the way described by the Bloomberg article" {
		t.Errorf("Incorrect response comment, should be '%s', was '%s'", "I’m assuming you mean specifically compromised in the way described by the Bloomberg article", response.Comment)
	}
}
