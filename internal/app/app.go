package app

import (
	"collyclicker/internal/Util"
	"collyclicker/internal/fileutils"
	"collyclicker/internal/scraper"
	"fmt"
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
	/*
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

		}*/
	fbref := []string{}
	c := colly.NewCollector(
		colly.AllowedDomains("fbref.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	c.IgnoreRobotsTxt = true

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://fbref.com/")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
	})

	// Inline tbody handler for one-time link scraping
	c.OnHTML("tbody", func(tbody *colly.HTMLElement) {
		fmt.Println("Found the tbody")
		fmt.Println(tbody.Text)
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
	fmt.Printf("Collected %d links:\n", len(fbref))
	for _, link := range fbref {
		fmt.Println(link)
	}
	fileutils.WriteCSVsingle("links.csv", fbref)
	Util.CheckURL("links.csv")

	// --- Add scraping state ---
	var keeperCounter int
	pageData := []scraper.TeamData{{}, {}}

	// --- Get handlers ---
	handlers := scraper.GetSelectorHandlers(&pageData, &keeperCounter, &fbref)

	// --- Register handlers with collector ---
	for _, h := range handlers {
		c.OnHTML(h.Selector, h.Handler)
	}
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s\n", r.Request.URL)
		fmt.Printf("Status Code: %d\n", r.StatusCode)
		fmt.Printf("Response Body: %s\n", string(r.Body))
		fmt.Printf("Error: %v\n", err)
	})

	return nil

}
