package predictions

import (
	"context"
	"golang.org/x/net/html"
	"strconv"
	"sync"
	"testing"
	"time"
)

type TestHtmlFetcher struct {
	GetHtmlFunc func(context.Context, string) (*html.Node, error)
}

func (f *TestHtmlFetcher) GetHtml(ctx context.Context, url string) (*html.Node, error) {
	return f.GetHtmlFunc(ctx, url)
}

func TestRetrievePredictionListPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			if url != "https://example.org/predictions/page/2" {
				t.Errorf("Incorrect response page requested, should be %s, was %s", "https://example.org/predictions/page/2", url)
			}

			return testHtmlLoad(t, "test_list.html"), nil
		},
	}

	s := NewSource(fetcher, "https://example.org")
	predictions, pageInfo, err := s.RetrievePredictionListPage(ctx, 2)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if pageInfo.LastPage != 287 {
		t.Errorf("Incorrect last page; should be %d, was %d", 287, pageInfo.LastPage)
	}
	if len(predictions) != 50 {
		t.Errorf("Incorrect number of prediction summaries; should be  %d, was %d", 50, len(predictions))
	}
	if predictions[1].Id != 193472 {
		t.Errorf("Second prediction had incorrect ID, should be %d, was %d", 193472, predictions[1].Id)
	}
}

func TestRetrievePredictionResponses(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			if url != "https://example.org/predictions/193436" {
				t.Errorf("Incorrect response page requested, should be %s, was %s", "https://example.org/predictions/193436", url)
			}

			return testHtmlLoad(t, "test_responses.html"), nil
		},
	}

	s := NewSource(fetcher, "https://example.org")
	responses, err := s.RetrievePredictionResponses(ctx, 193436)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if len(responses) != 8 {
		t.Errorf("Incorrect number of respoonses; should be %d, was %d", 8, len(responses))
	}
	if responses[1].Time.Unix() != 1539248198 {
		t.Errorf("Second prediction had incorrect time, should be %d, was %d", 1539248198, responses[1].Time.Unix())
	}
}

func TestLatest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			if url != "https://example.org/predictions/page/1" {
				t.Errorf("Incorrect response page requested, should be %s, was %s", "https://example.org/predictions/page/1", url)
			}

			return testHtmlLoad(t, "test_list.html"), nil
		},
	}

	s := NewSource(fetcher, "https://example.org")
	summary, err := s.Latest(ctx)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if summary.Id != 193473 {
		t.Errorf("Incorrect latest prediction ID; should be %d, was %d", 193473, summary.Id)
	}
}

func TestPredictionPageCount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			if url != "https://example.org/predictions/page/1" {
				t.Errorf("Incorrect response page requested, should be %s, was %s", "https://example.org/predictions/page/1", url)
			}

			return testHtmlLoad(t, "test_list.html"), nil
		},
	}

	s := NewSource(fetcher, "https://example.org")
	pageCount, err := s.PredictionPageCount(ctx)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if pageCount != 287 {
		t.Errorf("Incorrect page count; should be %d, was %d", 287, pageCount)
	}
}

func TestAllPredictions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	callCount := 0
	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			callCount++
			expectedUrl := "https://example.org/predictions/page/" + strconv.Itoa(callCount)
			if url != expectedUrl {
				t.Errorf("Incorrect response page requested, should be %s, was %s", expectedUrl, url)
			}

			if callCount != 287 {
				return testHtmlLoad(t, "test_list.html"), nil
			} else {
				return testHtmlLoad(t, "test_list_last.html"), nil
			}
		},
	}

	s := NewSource(fetcher, "https://example.org")
	summaries, err := s.AllPredictions(ctx)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if callCount != 287 {
		t.Errorf("GetHtml called incorrect number of times, should be %d, was %d", 287, callCount)
	}
	if len(summaries) != 81 {
		t.Errorf("Retrieved incorrect number of predictions; should be %d, was %d", 81, len(summaries))
	}
	if summaries[30].Id != 38 {
		t.Errorf("31st prediction had wrong ID; should be %d, was %d", 38, summaries[50].Id)
	}
}

func TestAllPredictionsSince(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	callCount := 0
	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			callCount++
			expectedUrl := "https://example.org/predictions/page/" + strconv.Itoa(callCount)
			if url != expectedUrl {
				t.Errorf("Incorrect response page requested, should be %s, was %s", expectedUrl, url)
			}

			if callCount == 1 {
				return testHtmlLoad(t, "test_list_empty.html"), nil
			} else if callCount != 287 {
				return testHtmlLoad(t, "test_list.html"), nil
			} else {
				return testHtmlLoad(t, "test_list_last.html"), nil
			}
		},
	}

	s := NewSource(fetcher, "https://example.org")
	summaries, err := s.AllPredictionsSince(ctx, time.Unix(1537670940, 0))
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if callCount != 2 {
		t.Errorf("GetHtml called incorrect number of times, should be %d, was %d", 2, callCount)
	}
	if len(summaries) != 46 {
		t.Errorf("Retrieved incorrect number of predictions; should be %d, was %d", 46, len(summaries))
	}
}

func TestAllPredictionResponses(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	summaries := []*PredictionSummary{
		{
			Id: 400,
		},
		{
			Id: 7,
		},
	}
	matchedSummaries := make(map[int64]bool)
	var mutex sync.Mutex

	callCount := 0
	fetcher := &TestHtmlFetcher{
		GetHtmlFunc: func(ctx context.Context, url string) (*html.Node, error) {
			mutex.Lock()
			defer mutex.Unlock()

			callCount++
			var match bool
			for _, summary := range summaries {
				if matchedSummaries[summary.Id] {
					continue
				}

				expectedUrl := "https://example.org/predictions/" + strconv.FormatInt(summary.Id, 10)
				if url == expectedUrl {
					matchedSummaries[summary.Id] = true
					match = true
					break
				}
			}

			if !match {
				t.Errorf("Incorrect response page requested,, was %s", url)
			}

			return testHtmlLoad(t, "test_responses.html"), nil
		},
	}

	s := NewSource(fetcher, "https://example.org")
	responses, err := s.AllPredictionResponses(ctx, summaries)
	if err != nil {
		t.Errorf("Error should have been nil, was %s", err)
	}
	if callCount != 2 {
		t.Errorf("GetHtml called incorrect number of times, should be %d, was %d", 2, callCount)
	}
	if len(responses) != 16 {
		t.Errorf("Retrieved incorrect number of responses; should be %d, was %d", 16, len(responses))
	}
	if responses[9].Prediction != 400 {
		t.Errorf("10th response had wrong prediction ID; should be %d, was %d", 400, responses[9].Prediction)
	}
	if responses[10].Time.Unix() != 1539316508 {
		t.Errorf("11th response had wrong time; should be %d, was %d", 1539316508, responses[9].Time.Unix())
	}
}
