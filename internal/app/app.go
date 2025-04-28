package app

import (
	"collyclicker/internal/csvparser"
	"collyclicker/internal/fileutils"
	"fmt"
	"io"
)

func Run() error {
	fmt.Println("Start scraping")
	tempfilename := "links/"

	csvArray, err := fileutils.Findcsvfiles(tempfilename)
	if err != nil {
		return err
	}

	for _, record := range csvArray {
		csvfile, err := csvparser.NewCSViter(fmt.Sprintf("links/" + record))
		if err != nil {
			return err
		}

		for {
			record, line, col, err := csvfile.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error at line %d : %w", line, err)
			}
			fmt.Printf("Line %d, Col %d: %v\n", line, col, record)
		}
		defer csvfile.Close()
	}
	return nil
}
