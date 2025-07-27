package Util

import (
	"fmt"
	"os"
	"path/filepath"
)

type RetryCache struct {
	url  string
	file string
}
type TrackCache struct {
	currentURL  string
	currentFile string
	index       int //Where in the current file is the current URL
}

func TmpDirCreate(name string) string {
	path := filepath.Join(os.TempDir(), name)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		Logger.Error("Failed to create temp directory", "Error", err)
	}
	return path
}

func TmpFileCreate(name string) string {
	path := filepath.Join(os.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		Logger.Error("Failed to create tmp file",
			"Error", err,
			"Location", "Cache.go : TmpFileCreate")
	}
	f.Close()
	return path
}

func OpenTempFile(name string) (*os.File, error) {
	path := filepath.Join(os.TempDir(), name)
	return os.Open(path)
}

func TruncateTmpFile(name string, tc TrackCache) error {
	path := filepath.Join(os.TempDir(), name)

	//truncate
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		Logger.Error("There was an error attempting to truncate the tmp file", "Error", err,
			"Location", "cache.go truncateTmpFile ")
		return err
	}
	defer f.Close()
	last := fmt.Sprintf("%s,%s", tc.currentFile, tc.currentURL)

	return nil
}

//check links/scrapeReady for retry.csv
//If
//Check links/ for LastURL.csv
//If LastURL exists or if os.tempdir exists -> Map entry look for link inside inside CSV file listed and start from there
//Maybe get index of entry from file Add function to csvParser
//Else Doesn't exist -> start fresh normal operation
