package predictions

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

type PredictionListPageInfo struct {
	Index    int64
	LastPage int64
}

func ExtractPredictionListPageInfo(rootNode *html.Node, index int64) (pageInfo *PredictionListPageInfo) {
	pageInfo = new(PredictionListPageInfo)
	pageInfo.Index = index

	page := goquery.NewDocumentFromNode(rootNode)

	lastPageHref, exists := page.Find("nav.pagination .last a").Attr("href")
	if exists {
		if strings.HasPrefix(lastPageHref, "/predictions/page/") {
			lastPageStr := lastPageHref[len("/predictions/page/"):]
			lastPage, err := strconv.ParseInt(lastPageStr, 10, 64)
			if err == nil {
				pageInfo.LastPage = lastPage
			}
		}
	} else {
		pageInfo.LastPage = index
	}

	return
}
