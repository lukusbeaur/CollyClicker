package app

import (
	"collyclicker/internal/Util"
	"collyclicker/internal/csvparser"
	"collyclicker/internal/fileutils"
	"collyclicker/internal/scraper"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var dirName_unready = "links/"
var dirname_ready = dirName_unready + "scrapeReady/"
var tempFilePath string

func Run() error {
	start := time.Now()
	Util.Logger.Debug("Starting Scraping process",
		"Location", "App.go Run()")
	//Create New collector
	//First c.OnHTML is pulling all the links with in the specified table. Reusing colly collector for everything else
	fbref := []string{}
	c := colly.NewCollector(
		colly.AllowedDomains("fbref.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	//ignore the robot.txt directions
	c.IgnoreRobotsTxt = true

	//On request set additional headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://fbref.com/")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		Util.Logger.Debug("Setting On request Information, Mainly Headers",
			"Location", "c.OnRequest")
	})
	//temporary debugging on failed requests.
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s\n", r.Request.URL)
		fmt.Printf("Status Code: %d\n", r.StatusCode)
		//fmt.Printf("Response Body: %s\n", string(r.Body))
		fmt.Printf("Error: %v\n", err)
	})
	// Inline tbody handler for one-time link scraping
	c.OnHTML("tbody", func(tbody *colly.HTMLElement) {
		//fmt.Println("Found the tbody")
		//fmt.Println(tbody.Text)

		tbody.ForEach("td[data-stat='score']", func(i int, links *colly.HTMLElement) {
			href := links.ChildAttr("a", "href")
			if href != "" {
				fbref = append(fbref, fmt.Sprintf("https://fbref.com%s", href))
			}
			/*Util.Logger.Debug("Gathering Links from main schedule page",
			"Location", "c.onHTML",
			"Selector", href)
			*/
		})

	})
	// This is used to get all links from the main schedule page
	err := c.Visit("https://fbref.com/en/comps/9/schedule/Premier-League-Scores-and-Fixtures")
	if err != nil {
		Util.Logger.Error("Could not connect to the link provided",
			slog.String("Location", "App.go - C.visit Error"),
			slog.Any("Error", err))
	}

	//Write all links scraped from FBref and save them/ If needed check to see if they are valid with Pinger
	fileutils.WriteCSVsingle(fmt.Sprintf(dirName_unready+"links.csv"), fbref)
	//Util.CheckURL("links/links.csv")

	//Find all CSVs in links folder stores them in a array
	//Inside the dirname_Ready folder will be all the vaid URLS for scraping.
	Util.Logger.Debug("Looking for CSV files in dirname Ready",
		slog.String("Location", "app.go - FindcsvFiles"),
		slog.Any("dirname_ready", dirname_ready))
	csvArray, err := fileutils.Findcsvfiles(dirname_ready)
	if err != nil {
		Util.Logger.Error("Trouble finding csvs in dirname_ready",
			slog.String("Location", "app.go - FindcsvFiles"),
			slog.Any("dirname_ready", dirname_ready),
			slog.Any("Error", err))
	}

	//Take array of CSV file names, and open one at a time.
	for _, record := range csvArray {
		Util.Logger.Debug("Reading CSVs from csvarray. Array created from findcsvfiles.",
			slog.String("Location", "app.go - Range csvArray loop"),
			slog.Any("Record", record))
		csvfile, err := csvparser.NewCSViter(fmt.Sprintf(dirname_ready + record))
		if err != nil {
			Util.Logger.Error("Trouble opening CSV file and or Iterator",
				slog.String("Location", "app.go - Range csvArray loop"),
				slog.Any("Record", record),
				slog.Any("Error", err))
		}
		defer csvfile.Close()

		for {
			// Start Function Time inside Outer CSV Loop -- Set Min max delay

			//TODO messed up delay function, moved it tou outer loop brain is fried it was inside handler loop but it delayed every handler
			//Need each URL to be Random not Each handler. Also the CSV is finished is inside the wrong loop too.
			funcDur := time.Now()

			minDelay := 2 * time.Second
			maxDelay := 15 * time.Second

			delay := minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay)))

			time.Sleep(delay)
			row, _, _, err := csvfile.Next()
			if errors.Is(err, io.EOF) {
				Util.Logger.Error("Error Opening CSV/ CSV is empty - CONTINUE",
					slog.String("Location", "app.go - Range csvArray loop -> inside For loop"),
					slog.Any("Row", row),
					slog.Any("Error", err))
				break
			} else if err != nil {
				Util.Logger.Error("Error Opening CSV/ Reading Row - keeps running",
					slog.String("Location", "app.go - Range csvArray loop -> inside For loop"),
					slog.Any("Row", row),
					slog.Any("Error", err))
				continue
			}
			//If starts with http its a url try and scrape with it.
			if len(row) == 0 || !strings.HasPrefix(row[0], "http") {
				continue
			}
			url := row[0]
			// Caching URL and file name
			curCache := &Util.TrackCache{
				CurrentURL:  url,
				CurrentFile: record,
				Index:       0,
				Sport:       "Soccer",
			}
			//Start Caching current links and URLS
			// Create a temporary directory / file for caching by sport handler
			tempFilePath, err = Util.CreateTempFile(*curCache)
			if err != nil {
				Util.Logger.Error("Error creating temporary file for caching",
					slog.String("Location", "app.go - Range csvArray loop -> inside For loop"),
					slog.Any("Error", err))
				continue
			}
			//truncate the temporary file to ensure it's empty before writing
			err = Util.TruncateTmpFile(*curCache)
			if err != nil {
				Util.Logger.Error("Error truncating temporary file for caching",
					"Error", err,
					"Location", "app.go - Range csvArray loop -> inside For loop")
				continue
			}
			// --- Add scraping state ---
			keeperCounter := 0
			pageData := []scraper.TeamData{{}, {}}
			Util.Logger.Info("Scraping State",
				slog.String("URL", url),
				slog.Int("Keeper Counter", keeperCounter))
			Util.Logger.Debug("Creation of CollyCollector : c:=")
			c := colly.NewCollector(
				colly.AllowedDomains("fbref.com"),
				colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
			)
			c.Limit((&colly.LimitRule{
				//For logging purpose I contorl the delay manually with in the main Loop.
				//RandomDelay: 10*time.Second + 5,
				DomainGlob: "*",
			}))
			//Ignore robot.txt and allow domain revisiting
			c.IgnoreRobotsTxt = true
			c.AllowURLRevisit = true
			//On request set additional headers
			c.OnRequest(func(r *colly.Request) {
				r.Headers.Set("Referer", "https://fbref.com/")
				r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
			})
			c.OnError(func(r *colly.Response, err error) {
				Util.Logger.Error("Colly request error",
					"url", r.Request.URL.String(),
					"status", r.StatusCode,
					"err", err,
				)

			})

			// --- Get handlers ---
			handlers := scraper.GetSelectorHandlers(&pageData, &keeperCounter, &fbref)

			dateStr, err := fileutils.ExtractDateFromURL(url)
			Util.Logger.Debug("Extracting Date from URL string for building output folder structure.",
				"Date", dateStr,
				"Location", "App.go - After Handler selectors -> Get Dates")
			if err != nil {
				Util.Logger.Error("Error Extracting date from ",
					"Location", "App.go - After Handler selectors -> Get Dates",
					"URL", url,
					"Error", err)
			}

			// --- Register handlers with collector ---
			for _, h := range handlers {
				handlerName := h.Name
				Util.Logger.Debug("Executing handler",
					slog.String("handler", handlerName),
					slog.String("selector", h.Selector),
					slog.String("location", "app.go - handler wrapper"))

				c.OnHTML(h.Selector, h.Handler)

				//Random Delay Configuraiton / Logging //

				Util.Logger.Info("Assigned scrape delay per URL",
					"delay", delay,
					"Scrape time", time.Since(funcDur),
					"url", url,

					"location", "app.go - scrape loop",
				)
				//-------------------------------------------------------------------//
				err = c.Visit(url)
				if err != nil {
					Util.Logger.Error("Error Extracting date from ",
						"Location", "App.go - After Handler selectors -> Get Dates",
						"URL", url,
						"Error", err)
					continue
				}

				Util.Logger.Debug("writing data to CSV",
					"Location", "PageDataTOCSV",
					//"PageData", pageData,
					"datestr", dateStr,
					"Location", "App.go - End of Inner most scraping loop  - Per URL")
				scraper.PageDataToCSV(pageData, dateStr)
				t := time.Now()
				elapsed := t.Sub(start)
				Util.Logger.Info("Finished scraping URL,",
					"Total Duration", elapsed,
					"Scrape time", time.Since(funcDur),
					"URL", url)
			}
			t := time.Now()
			elapsed := t.Sub(start)
			Util.Logger.Info("Finished CSV,",
				"Total Duration", elapsed,
				"Record", record)
		}

	}
	t := time.Now()
	elapsed := t.Sub(start)
	Util.Logger.Info("Finished scraping all URLS/ All CSVs.,",
		"Total Duration", elapsed)

	//Delete all temp files inside Temp DIR
	err = os.Remove(tempFilePath)
	if err != nil {
		slog.Error("Error Deleting Temp DIR. Consider checking your Temp folder for manual deletion",
			"Error", err,
			"Location", "App.go : Final Line")
	}
	return nil

}
