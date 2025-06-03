package fileutils

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func WriteCSVsingle(file string, links []string) {
	csvFile, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	defer w.Flush()
	for _, link := range links {
		if err := w.Write([]string{link}); err != nil {
			log.Fatalf("Error writing to file: %v ", err)
		}
	}
	fmt.Printf("Writing  %s to CSV Complete.\n ", file)

}
func WriteLineCSV(file string, link []string) {
	csvfile, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(csvfile)
	defer w.Flush()
	w.Write(link)
	fmt.Printf("Wrote line to %s \n", file)
}
