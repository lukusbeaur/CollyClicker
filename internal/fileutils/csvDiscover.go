// /fileutils/csvDiscover.go

package fileutils

import (
	"os"
	"strings"
)

// Open the specified directory and search for all csv files
// isDir is used to passover directorys/ folders, and the has suffix checks for csvs
func Findcsvfiles(path string) ([]string, error) {
	//does a CSV file exist at the path
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
