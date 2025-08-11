package Util

import (
	"bufio"
	"collyclicker/internal/fileutils"
	"fmt"
	"os"
	"path/filepath"
)

type TrackCache struct {
	Sport       string //sport type
	CurrentURL  string
	CurrentFile string
	Index       int //Where in the current file is the current URL
}

var TempFolder = "CollyClicker"

// Create a temporary directory in the system's temp directory
/*func TmpDirCreate(name string) (string, error) {
	path := filepath.Join(os.TempDir(), name)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		//Logger.Error("Failed to create temp directory", "Error", err)
		return "", err
	}
	return path, nil
}*/
func CreateTempFile(tc TrackCache) (string, error) {
	tempDir := filepath.Join(os.TempDir(), TempFolder)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}
	tempFilePath := filepath.Join(tempDir, tc.Sport)
	f, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	f.Close()
	return tempFilePath, nil
}

// TmpFileCreate creates a temporary file in the systems temp directory with the given function name
/*func TmpFileCreate(Sport) (string, error) {
	path := filepath.Join(os.TempDir(), Sport)
	f, err := os.Create(path)
	if err != nil {
		/*Logger.Error("Failed to create tmp file",
		"Error", err,
		"Location", "Cache.go : TmpFileCreate")*\

		return "", err
	}
	f.Close()
	return path, nil
}
*/

func OpenTempFile(name string) (*os.File, error) {
	path := filepath.Join(os.TempDir(), name)
	return os.Open(path)
}

// TruncateTmpFile truncates the temporary file to only contain the Last url and file name
func TruncateTmpFile(tc TrackCache) error {
	path := filepath.Join(os.TempDir(), TempFolder, tc.Sport)

	//truncate
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		/*Logger.Error("There was an error attempting to truncate the tmp file", "Error", err,
		"Location", "cache.go truncateTmpFile ")*/
		return err
	}
	defer f.Close()
	last := fmt.Sprintf("%s,%s", tc.CurrentFile, tc.CurrentURL)

	f.WriteString(last)
	return err
}

// GetIndex retrieves the index of a specific link in a file.
// it Returns the index of the link if found, otherwise returns -1.
func GetIndex(tc TrackCache) int {
	//instatiate tc.index
	tc.Index = 0
	f, err := os.Open(tc.CurrentFile)
	if err != nil {
		/*Logger.Error("Error opening file for index retreival",
		"Error", err,
		"Location", "cache.go GetIndex")
		*/
		return -1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == tc.CurrentURL {
			return tc.Index
		}
		tc.Index++
	}
	if err := scanner.Err(); err != nil {
		/*
			Logger.Error("Error reading file/lines for index retreival",
				"Error", err,
				"Location", "cache.go GetIndex")
			return -1*/
	}
	return -1
}

func AddToRetryCache(file string, url string) error {
	retryPath := "ScrapeReady/retryCache.csv"
	return fileutils.WriteLineCSV(retryPath, []string{file, url})
}

//check links/scrapeReady for retry.csv
//If
//Check links/ for LastURL.csv
//If LastURL exists or if os.tempdir exists -> Map entry look for link inside inside CSV file listed and start from there
//Maybe get index of entry from file Add function to csvParser
//Else Doesn't exist -> start fresh normal operation
