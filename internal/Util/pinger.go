package Util

import (
	"collyclicker/internal/fileutils"
	"encoding/csv"
	"fmt"
	"os"
)

// Take CSV file, open it. Create new CSV reader. Each URL is sent to sping.go checking for Status 200
// If status 200 -> write string to pass.csv If status != 200 -> write string to fail.csv
func CheckURL(file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	r := csv.NewReader(f)
	urls, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for _, url := range urls {
		//skip empty lines
		if len(url) == 0 {
			continue
		}
		row := url[0]

		if code, err := Ping(row); err != nil {
			panic(err)
		} else if code == 200 {
			fmt.Printf("Code:%d \n Link: %s\n", code, row)
			//fail = append(pass, row)
			fileutils.WriteLineCSV("links/scrapeReady/pass_CSV.csv", []string{row})
		} else {
			fmt.Printf("Code:%d \n Link: %s\n", code, row)
			//pass = append(fail, row)
			fileutils.WriteLineCSV("links/fail_CSV.csv", []string{row})
		}
	}
}
