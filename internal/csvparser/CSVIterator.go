// collyclicker/csvparser/csviterator.go
package csvparser

import (
	"encoding/csv"
	"os"
)

// Encapsulate everything needed inside csvIter for openign and reading a csv file
// removes the need for main to manage these items seperatly
type csvIter struct {
	file      *os.File
	reader    *csv.Reader
	recordNum int
}

// Opens CSV, and returns Iterator (CSV.NewReader)
func NewCSViter(path string) (*csvIter, error) {
	f, err := os.Open(path)
	if err != nil {
		//log.Fatalf("Error opening file %v", path)
		return nil, err
	}
	return &csvIter{
		file:      f,
		reader:    csv.NewReader(f),
		recordNum: 0,
	}, nil
}

func (it *csvIter) Next() ([]string, int, int, error) {
	//return a slice string from the CSV file
	//the line of the current record
	//the position of the column
	//error - obv

	record, err := it.reader.Read()
	if err != nil {
		return nil, 0, 0, err
	}
	//field position is tracked internally and Fieldpos(0) is saying hey start at the first posiiton in the new row
	//and it still returns the updated line
	line, column := it.reader.FieldPos(0)
	return record, line, column, nil
}

func (it *csvIter) Close() error {
	return it.file.Close()
}
