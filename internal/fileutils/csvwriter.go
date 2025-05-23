package fileutils

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
)

// WriteCSV writes the given headers and rows to a CSV file at the specified filepath.
// It ensures the directory exists, creates the file, and writes the data.
func WriteCSV(folderPath, fileName string, headers []string, rows [][]string) error {
	// Print working directory for context
	cwd, err := os.Getwd()
	if err == nil {
		log.Printf("📁 Current Working Directory: %s", cwd)
	} else {
		log.Printf("⚠️ Failed to get working directory: %v", err)
	}

	// Make sure the folder exists
	err = EnsureDir(folderPath)
	if err != nil {
		return err
	}

	// Create or truncate the CSV file
	fullPath := filepath.Join(folderPath, fileName)
	file, err := CreateCSVFile(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the headers first
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write all the rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	log.Printf("✅ CSV written: %s", fullPath)
	return nil
}
