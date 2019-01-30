package predictions

import (
	"math"
	"testing"
)

func TestExtractPredictionSummaryResponsePage(t *testing.T) {
	t.Parallel()

	rootNode := testHtmlLoad(t, "test_responses.html")
	prediction := ExtractPredictionSummaryResponsePage(123, rootNode)

	if prediction.Outcome != Right {
		t.Errorf("Incorrect prediction outcome, should be %d, was %d", Right, prediction.Outcome)
	}
	if prediction.Id != 123 {
		t.Errorf("Incorrect prediction ID, should be %d, was %d", 123, prediction.Id)
	}
	if prediction.Title != "Convincing evidence will prove that Amazon, Apple, etc.'s chips were compromised." {
		t.Errorf("Incorrect prediction title, should be %s, was %s", "Convincing evidence will prove that Amazon, Apple, etc.'s chips were compromised.", prediction.Title)
	}
	if prediction.Creator != "krazemon" {
		t.Errorf("Incorrect prediction creator, should be %s, was %s", "krazemon", prediction.Creator)
	}
	if prediction.Created.Unix() != 1538836809 {
		t.Errorf("Incorrect prediction created time, should be %d, was %d", 1548870174, prediction.Created.Unix())
	}
	if prediction.Deadline.Unix() != 1554552000 {
		t.Errorf("Incorrect prediction deadline, should be %d, was %d", 1541088000, prediction.Deadline.Unix())
	}
	if math.Abs(prediction.MeanConfidence-0.25857142857) > 0.00001 {
		t.Errorf("Incorrect prediction mean confidence, should be %g, was %g", 0.25857142857, prediction.MeanConfidence)
	}
	if prediction.WagerCount != 8 {
		t.Errorf("Incorrect prediction wager count, should be %d, was %d", 8, prediction.WagerCount)
	}
}
