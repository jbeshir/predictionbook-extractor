package predictions

import (
	"context"
	"golang.org/x/net/html"
	"testing"
)

type TestHtmlFetcher struct {
	GetHtmlFunc func(context.Context, string) (*html.Node, error)
}

func (f *TestHtmlFetcher) GetHtml(ctx context.Context, url string) (*html.Node, error) {
	return f.GetHtmlFunc(ctx, url)
}

func TestRetrievePredictionListPage(t *testing.T) {

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
