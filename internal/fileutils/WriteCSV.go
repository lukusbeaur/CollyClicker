package fileutils

import (
	"encoding/csv"
	"os"
)

// WriteCSVsingle creates a new CSV file and writes a slice of strings (links) to it
// truncates the file if it already exists
func WriteCSVsingle(file string, links []string) error {
	csvFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	defer w.Flush()
	for _, link := range links {
		if err := w.Write([]string{link}); err != nil {
			return err
		}
	}
	return nil
}

// WriteLineCSV appends a line to a CSV file
// it takes a file and a slice of strings ( link ) to write as a new line
func WriteLineCSV(file string, link []string) error {
	csvfile, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	defer w.Flush()
	if err := w.Write(link); err != nil {
		return err
	}
	return nil
}
