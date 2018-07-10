package main

import (
	"flag"
	"github.com/jbeshir/predictionbook-extractor/predictions"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"golang.org/x/time/rate"
)

func main() {

	url := flag.String("url", "https://predictionbook.com", "URL of PredictionBook instance to extract from")
	flag.Parse()

	extractor := htmlextract.NewExtractor(rate.NewLimiter(1, 2))
	predictions.NewSource(extractor, *url)
}
