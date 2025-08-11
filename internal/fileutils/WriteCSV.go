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
	//Move to where you call this function..
	/* EXAMPLE USAGE
		err := fileutils.WriteCSVsingle(file, links)
	if err != nil {
	    Logger.Error("Failed to write CSV", "Error", err)
	} else {
	    Logger.Info(fmt.Sprintf("Writing %s to CSV Complete.", file))
	}
	*/
	// fmt.Printf("Writing  %s to CSV Complete.\n ", file)
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
	//Check WriteCSVsingle for example usage
	//fmt.Printf("Wrote line to %s \n", file)
	return nil
}
