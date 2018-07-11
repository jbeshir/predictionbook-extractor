package predictions

import (
	"strconv"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"time"
	"context"
	"golang.org/x/net/html"
	"errors"
	"strings"
)

type Source struct {
	baseUrl string
	extractor *htmlextract.Extractor
}

func NewSource(extractor *htmlextract.Extractor, baseUrl string) *Source {
	return &Source{
		baseUrl: baseUrl,
		extractor: extractor,
	}
}

func (s* Source) Latest(ctx context.Context) (*PredictionSummary, error) {
	latest, _, err := s.retrievePredictionPage(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(latest) == 0 {
		return nil, errors.New("no predictions found")
	}

	return latest[0], nil
}

func (s *Source) PredictionPageCount(ctx context.Context) (int64, error) {
	_, info, err := s.retrievePredictionPage(ctx, 1)
	if err != nil {
		return 0, err
	}
	if info.LastPage == 0 {
		return 0, errors.New("unable to extract page count")
	}

	return info.LastPage, nil
}

func (s* Source) retrievePredictionPage(ctx context.Context, index int64) (predictions []*PredictionSummary, pageInfo *predictionPageInfo, err error) {
	page, err := s.extractor.GetHtml(ctx, s.baseUrl + "/predictions/page/" + strconv.FormatInt(index, 10))
	if err != nil {
		return nil, nil, err
	}

	predictionNodes := htmlextract.HtmlNodesByAttr(page, "", "", "class", "prediction")
	for _, node := range predictionNodes {
		prediction := new(PredictionSummary)

		creatorNodes := htmlextract.HtmlNodesByAttr(node, "", "", "class", "creator")
		if len(creatorNodes) > 0 && creatorNodes[0].FirstChild != nil && creatorNodes[0].FirstChild.Type == html.TextNode {
			prediction.Creator = creatorNodes[0].FirstChild.Data
		}

		createdAtNodes := htmlextract.HtmlNodesByAttr(node, "", "", "class", "created_at")
		if len(createdAtNodes) > 0 {
			for _, attr := range createdAtNodes[0].Attr {
				if attr.Key == "title" {
					createdAt, err := time.Parse("2006-01-02 15:04:05 MST", attr.Val)
					if err == nil {
						prediction.Created = createdAt
					}
					break
				}
			}
		}

		predictions = append(predictions, prediction)
	}

	pageInfo = new(predictionPageInfo)
	pageInfo.Index = index

	paginationNodes := htmlextract.HtmlNodesByAttr(page, "", "", "class", "pagination")
	if len(paginationNodes) > 0 {
		lastPageNodes := htmlextract.HtmlNodesByAttr(paginationNodes[0], "", "", "class", "last")
		if len(lastPageNodes) > 0 {
			linkNodes := htmlextract.HtmlNodesByAttr(lastPageNodes[0], "a", "", "", "")
			if len(linkNodes) > 0 {
				for _, attr := range linkNodes[0].Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "/predictions/page/") {
							lastPageStr := attr.Val[len("/predictions/page/"):]
							lastPage, err := strconv.ParseInt(lastPageStr, 10, 64)
							if err == nil {
								pageInfo.LastPage = lastPage
							}
						}
					}
				}
			}
		} else {
			pageInfo.LastPage = index
		}
	}


	return
}

type PredictionSummary struct {
	Id int64
	Text string
	Creator string
	Created time.Time
	Deadline time.Time
	MeanConfidence float64
}

type predictionPageInfo struct {
	Index int64
	LastPage int64
}