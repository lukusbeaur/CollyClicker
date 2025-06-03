package Util

import (
	"encoding/csv"
	"fmt"
	"os"

	"collyclicker/internal/fileutils"
)

func CheckURL(file string) {
	var pass []string
	var fail []string
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
			fileutils.WriteLineCSV("pass_CSV.csv", []string{row})
		} else {
			fmt.Printf("Code:%d \n Link: %s\n", code, row)
			//pass = append(fail, row)
			fileutils.WriteLineCSV("fail_CSV.csv", []string{row})
		}
	}
	fileutils.WriteCSVsingle("fail_CSV.csv", fail)
	fileutils.WriteCSVsingle("pass_CSV.csv", pass)

}
