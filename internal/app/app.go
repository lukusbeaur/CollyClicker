package app

import (
	"collyclicker/internal/Util"
	"collyclicker/internal/csvparser"
	"collyclicker/internal/fileutils"
	"fmt"
	"io"
	"log"

	"github.com/gocolly/colly/v2"
)

func Run() error {
	fmt.Println("Start scraping")
	tempfilename := "links/"

	csvArray, err := fileutils.Findcsvfiles(tempfilename)
	if err != nil {
		return err
	}

	for _, record := range csvArray {
		csvfile, err := csvparser.NewCSViter(fmt.Sprintf("links/" + record))
		if err != nil {
			return err
		}

		for {
			record, line, col, err := csvfile.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error at line %d : %w", line, err)
			}
			fmt.Printf("Line %d, Col %d: %v\n", line, col, record)
		}
		defer csvfile.Close()
	}
	fbref := []string{}
	c := colly.NewCollector(
		colly.AllowedDomains("fbref.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	//Get table
	c.OnHTML("tbody", func(tbody *colly.HTMLElement) {
		fmt.Println("Found the tbody")
		//fmt.Println(tbody.Text)
		//tbody.ChildAttr(`td[data-stat="score"] a`, "href")
		tbody.ForEach("td[data-stat='score']", func(i int, links *colly.HTMLElement) {
			href := links.ChildAttr("a", "href")

			if href != "" {
				fbref = append(fbref, fmt.Sprintf("https://fbref.com%s", href))
			}
		})
	})

	err = c.Visit("https://fbref.com/en/comps/9/schedule/Premier-League-Scores-and-Fixtures")
	if err != nil {
		log.Fatalf("Could not connect to the link provided: %v", err)
	}
	fileutils.WriteCSVsingle("links.csv", fbref)
	Util.CheckURL("links.csv")

	return nil

}
