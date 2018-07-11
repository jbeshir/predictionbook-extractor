package main

import (
	"flag"
	"github.com/jbeshir/predictionbook-extractor/predictions"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"golang.org/x/time/rate"
	"context"
	"fmt"
)

func main() {

	url := flag.String("url", "https://predictionbook.com", "URL of PredictionBook instance to extract from")
	flag.Parse()

	extractor := htmlextract.NewExtractor(rate.NewLimiter(1, 2), 2)
	source := predictions.NewSource(extractor, *url)

	latest, err := source.Latest(context.Background())
	if err != nil {
		fmt.Errorf("Error retrieving latest prediction: %s\n", err)
		return
	}

	pageCount, err := source.PredictionPageCount(context.Background())
	if err != nil {
		fmt.Errorf("Error retrieving prediction page count: %s\n", err)
		return
	}

	fmt.Println(latest.Created.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Pages of predictions: %d", pageCount)
}
