// /fileutils/csvDiscover.go

package fileutils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Open the specified directory and search for all csv files
// isDir is used to passover directorys/ folders, and the has suffix checks for csvs
func Findcsvfiles(path string) ([]string, error) {
	//does a CSV file exist at the path)
	csvLists := []string{}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !(entry.IsDir()) && strings.HasSuffix(entry.Name(), ".csv") {
			csvLists = append(csvLists, entry.Name())
		}

	}
	return csvLists, nil
}

func ReadLinksFromCSV(filePath string) ([]string, error) {
	var links []string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(bufio.NewReader(file))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range records {
		for _, entry := range row {
			entry = strings.TrimSpace(entry)
			if entry != "" {
				links = append(links, entry)
			}
		}
	}

	return links, nil
}

// Extract month-day-year from link
func ExtractDateFromURL(url string) (string, error) {
	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	//fmt.Println("Current URL:", url)
	for _, month := range months {
		// Check if the month is present in the URL
		if idx := strings.Index(url, month); idx != -1 {
			// Get substring starting from month
			sub := url[idx:]
			parts := strings.Split(sub, "-")
			if len(parts) >= 3 {
				//
				dateStr := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
				return dateStr, nil
			}
		}
	}
	//Log an error in App if no date is found
	return "", nil
}
