package predictions

import (
	"strconv"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"time"
	"context"
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

func (s* Source) MostRecent() {

}

func (s* Source) retrievePredictionPage(ctx context.Context, index int64) (predictions []PredictionSummary, err error) {
	url := s.baseUrl + "/page/" + strconv.FormatInt(index, 64)
	_, err = s.extractor.HtmlNodesByAttrInPages(ctx, []string{url}, "", "class", "prediction")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type PredictionSummary struct {
	id int64
	text string
	author string
	created time.Time
}