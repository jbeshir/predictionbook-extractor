package predictions

import (
	"golang.org/x/net/html"
	"os"
	"path/filepath"
	"testing"
)

func testHtmlLoad(t *testing.T, name string) (rootNode *html.Node) {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("helperListLoad: Couldn't open test list page: %s", err)
	}

	rootNode, err = html.Parse(f)
	if err != nil {
		t.Fatalf("helperListLoad: Couldn't parse test list page: %s", err)
	}

	return
}

func TestExtractPageInfo(t *testing.T) {
	rootNode := testHtmlLoad(t, "test_list.html")

	pageInfo := ExtractPredictionListPageInfo(rootNode, 2)
	if pageInfo.Index != 2 {
		t.Errorf("Incorrect page index; should be %d, was %d", 2, pageInfo.Index)
	}
	if pageInfo.LastPage != 287 {
		t.Errorf("Incorrect page index; should be %d, was %d", 287, pageInfo.LastPage)
	}
}

func TestExtractPageInfoLast(t *testing.T) {
	rootNode := testHtmlLoad(t, "test_list_last.html")

	pageInfo := ExtractPredictionListPageInfo(rootNode, 287)
	if pageInfo.Index != 287 {
		t.Errorf("Incorrect page index; should be %d, was %d", 287, pageInfo.Index)
	}
	if pageInfo.LastPage != 287 {
		t.Errorf("Incorrect page index; should be %d, was %d", 287, pageInfo.LastPage)
	}
}
