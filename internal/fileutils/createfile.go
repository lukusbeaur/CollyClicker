package fileutils

import (
	"os"
)

// ensureDir checks if a folder exists at the given path, and creates it if missing.
func EnsureDir(path string) error {

	return os.MkdirAll(path, os.ModePerm)
}

// createCSVFile creates (or overwrites) a CSV file at the given path and returns the *os.File.
func CreateCSVFile(filepath string) (*os.File, error) {

	return os.Create(filepath)
}
