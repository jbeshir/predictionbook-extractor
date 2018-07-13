package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/jbeshir/predictionbook-extractor/htmlextract"
	"github.com/jbeshir/predictionbook-extractor/predictions"
	"golang.org/x/time/rate"
	"os"
	"strconv"
)

func main() {

	url := flag.String("url", "https://predictionbook.com", "URL of PredictionBook instance to extract from")
	export := flag.String("export", "", "Export all predictions made in CSV format to the given file")
	flag.Parse()

	extractor := htmlextract.NewExtractor(rate.NewLimiter(1, 2), 2)
	source := predictions.NewSource(extractor, *url)

	if *export != "" {
		exportFile, err := os.Create(*export)
		if err != nil {
			fmt.Errorf("Error opening export file: %s\n", err)
			return
		}

		ps, err := source.AllPredictions(context.Background())
		if err != nil {
			fmt.Errorf("Error retrieving predictions: %s\n", err)
			return
		}

		csvWriter := csv.NewWriter(exportFile)
		for _, p := range ps {
			err := csvWriter.Write([]string{
				strconv.FormatInt(p.Id, 10),
				strconv.FormatInt(p.Created.Unix(), 10),
				strconv.FormatInt(p.Deadline.Unix(), 10),
				strconv.FormatFloat(p.MeanConfidence, 'f', -1, 64),
				strconv.FormatInt(p.WagerCount, 10),
				strconv.FormatInt(int64(p.Outcome), 10),
				p.Creator,
				p.Title,
			})
			if err != nil {
				fmt.Errorf("Error writing predictions: %s\n", err)
				return
			}
		}

		csvWriter.Flush()
		err = csvWriter.Error()
		if err != nil {
			fmt.Errorf("Error writing predictions: %s\n", err)
			return
		}

		err = exportFile.Close()
		if err != nil {
			fmt.Errorf("Error writing predictions: %s\n", err)
			return
		}
	}
}
