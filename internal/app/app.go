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
	"net/http"
	"os"
	"strconv"
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
	//On request set additional headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://fbref.com/")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
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
	Util.Logger.Info("Looking for CSV files in dirname Ready",
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
			// Start Function Time inside Outer CSV Loop
			funcDur := time.Now()
			row, line, _, err := csvfile.Next()
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
				Index:       line,
				Sport:       "Soccer",
				CacheType:   "current",
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

			//Set up the collector with the necessary configurations--------------------->
			c := colly.NewCollector(
				colly.AllowedDomains("fbref.com"),
				colly.Async(false), // keep it simple while debugging
				colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
			)
			c.Limit(&colly.LimitRule{
				DomainGlob:  "*",
				Parallelism: 1,
				Delay:       2 * time.Second, // min
				RandomDelay: 5 * time.Second, // extra random, so 2–15s total
			})

			// Dont ignore robot.txt but allow domain revisiting
			c.IgnoreRobotsTxt = false
			c.AllowURLRevisit = true
			//On request set additional headers

			const maxRetries = 2
			c.OnRequest(func(r *colly.Request) {
				r.Headers.Set("Referer", "https://fbref.com/")
				if _, ok := r.Ctx.GetAny("retryCount").(int); !ok {
					r.Ctx.Put("retryCount", 0)
				}
			})
			c.OnError(func(r *colly.Response, err error) {

				if r == nil {
					// network or other error before response
					Util.Logger.Error("Request failed before response", "err", err)
					return
				}

				if r.StatusCode != 429 {
					Util.Logger.Error("Colly request error",
						"url", r.Request.URL.String(),
						"status", r.StatusCode,
						"err", err,
					)
					return
				}
				Util.Logger.Error("Colly request error",
					"url", r.Request.URL.String(),
					"status", r.StatusCode,
					"err", err,
				)
				// r.ctx is a per request context, so we can store retry count
				// "retryCount" is a key in the context to track retries
				// .(int) is a type assertion to get the retry count
				// no need for global variable, as each request has its own context
				// no race condition, as each request has its own context
				retryCount := r.Ctx.GetAny("retryCount").(int)
				if retryCount >= maxRetries {
					Util.Logger.Error("429 max retries reached",
						"url", r.Request.URL.String(),
						"retries", retryCount,
					)
					// If max retries reached, add URL and filename to cache
					Util.AddToRetryCache(url, record)
					Util.Logger.Warn("Added to retry cache",
						"url", url,
						"filename", record,
						"Location", "App.go - OnError 429 handler")
					return
				}
				// Compute sleep from Retry-After (seconds or HTTP-date), else default
				sleep := 60 * time.Second
				// If Retry-After header is present, use it to determine sleep duration
				// strconv.Atoi converts string to int, http.ParseTime parses HTTP-date
				//if e == no error than seconds is the right format
				// if e != nil then it is not seconds, so try to parse it as HTTP-date
				if ra := r.Headers.Get("Retry-After"); ra != "" {
					if secs, e := strconv.Atoi(strings.TrimSpace(ra)); e == nil {
						sleep = time.Duration(secs) * time.Second
					} else if when, e := http.ParseTime(ra); e == nil {
						if d := time.Until(when); d > 0 && d < 30*time.Minute {
							sleep = d
						}
					}
				}

				// Exponential backoff with jitter
				// 1<<retryCount is a bit shift operation that calculates 2^retryCount
				// This gives us 1, 2, 4, 8, ... seconds for each retry
				// We cap the backoff at 60 seconds to avoid excessive delays
				backoff := time.Duration(1<<retryCount) * time.Second // 1s,2s,4s,8s...
				if backoff > 60*time.Second {
					backoff = 60 * time.Second
				}
				jitter := time.Duration(rand.Intn(4000)) * time.Millisecond
				wait := sleep + backoff + jitter

				Util.Logger.Warn("429 received; backing off",
					"url", r.Request.URL.String(),
					"retryAfter", sleep,
					"backoff", backoff,
					"wait", wait,
					"retries", retryCount+1,
				)

				time.Sleep(wait)
				r.Ctx.Put("retryCount", retryCount+1)
				_ = r.Request.Retry()

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

				//c.OnHTML(h.Selector, h.Handler)
				c.OnHTML(h.Selector, func(e *colly.HTMLElement) {
					startHandler := time.Now()
					h.Handler(e) //  handler logic
					elapsedHandler := time.Since(startHandler)
					Util.Logger.Info("Handler finished",
						"Selector", h.Selector,
						"Duration", elapsedHandler,
						"URL", e.Request.URL.String())
				})
				//Random Delay Configuraiton / Logging //

				Util.Logger.Info("Total Scrape time",
					"Scrape time", time.Since(funcDur),
					"url", url,
					"location", "app.go - scrape loop",
				)
				//-------------------------------------------------------------------//

				scraper.PageDataToCSV(pageData, dateStr)
				t := time.Now()
				elapsed := t.Sub(start)
				Util.Logger.Info("Finished scraping Handler ,",
					"Total Duration", elapsed,
					"Scrape time", time.Since(funcDur),
					//"Handler", handlers)
				)
			}
			//--- Visit URL and log the time right before and after for analysis ---
			// This is where the actual scraping happens
			urlStart := time.Now()
			err = c.Visit(url)
			t := time.Now()
			Util.Logger.Info("Finished URL Visit",
				"URL", url,
				"Duration", t.Sub(urlStart),
				"Location", "App.go - c.Visit()")
			if err != nil {
				Util.Logger.Error("Error Extracting date from ",
					"Location", "App.go - c.Visit()",
					"URL", url,
					"Error", err)
				continue
			}
			t = time.Now()
			elapsed := t.Sub(start)
			Util.Logger.Info("Finished URL,",
				"Total Duration", elapsed,
				"URL", url)
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
