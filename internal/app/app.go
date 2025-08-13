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

// Setting global variabl Max retries for 429 errors
const maxRetries = 2

//var url string

func Run() error {
	start := time.Now()
	Util.Logger.Debug("Starting Scraping process",
		"Location", "App.go Run()")
	//Create New collector
	//First c.OnHTML is pulling all the links with in the specified table. Reusing colly collector for everything else
	//This initial scrape does not work , fbref removed all 'score links' from the main schedule page.
	fbref := []string{}

	// --- Colly Collector set up Non call back handlers ------------------------------>
	c := colly.NewCollector(
		colly.AllowedDomains("fbref.com"),
		colly.Async(false),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       2 * time.Second, // min
		RandomDelay: 4 * time.Second, // extra random
	})
	// Dont ignore robot.txt but allow domain revisiting
	c.IgnoreRobotsTxt = false
	c.AllowURLRevisit = true
	//------------------end of Non callback handlers---------------------------------------->
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
	//I dont feel like refactoring this right now so im just going to skip it with this.
	var links = false
	if links {
		err := c.Visit("https://fbref.com/en/comps/9/schedule/Premier-League-Scores-and-Fixtures")
		if err != nil {
			Util.Logger.Error("Could not connect to the link provided",
				slog.String("Location", "App.go - C.visit Error"),
				slog.Any("Error", err))
		}

		//Write all links scraped from FBref and save them/ If needed check to see if they are valid with Pinger
		fileutils.WriteCSVsingle(fmt.Sprintf(dirName_unready+"links.csv"), fbref)
		//Util.CheckURL("links/links.csv")
	}
	//Find all CSVs in links folder stores them in a array
	//Inside the dirname_Ready folder will be all the vaid URLS for scraping.
	Util.Logger.Info("Looking for CSV files in dirname Ready",
		"Location", "app.go - FindcsvFiles",
		"dirname_ready", dirname_ready)

	csvArray, err := fileutils.Findcsvfiles(dirname_ready)
	if err != nil {
		Util.Logger.Error("Trouble finding csvs in dirname_ready",
			"Location", "app.go - FindcsvFiles",
			"dirname_ready", dirname_ready,
			"Error", err)
	}

	Util.Logger.Debug("Creation of CollyCollector : c:=")

	//-----------------get Caching info ----------------------------------------->
	cacheInfo, err := Util.OpenTempFileString("Soccer")
	if err != nil {
		Util.Logger.Error("Error opening temporary file for caching",
			"Location", "app.go - OpenTempFileString",
			"Sport", "Soccer",
			"Error", err)
		cacheInfo = []string{}
	}
	var cacheFile string
	var cacheIndexInt = -1
	if len(cacheInfo) >= 3 {
		cacheFile = cacheInfo[0]
		if ci, e := strconv.Atoi(cacheInfo[2]); e == nil {
			cacheIndexInt = ci
		}
	}
	Util.Logger.Debug("Cache file and index",
		"Index", cacheIndexInt,
		"CacheFile", cacheFile,
		"Location", "app.go - OpenTempFileString")
	//-----------------end of Caching info --------------------------------------->

	//Take array of CSV file names, and open one at a time.
	for _, record := range csvArray {

		//Check for cache info and if it exists, skip the file-------------->
		if record != cacheFile && cacheFile != "" {
			continue // skip this file if it is not the cache file
		}
		//Caching logic over                   ------------------------------>
		Util.Logger.Debug("Reading CSVs from csvarray.",
			"Location", "app.go - Range csvArray loop",
			"Record", record)

		// Create a new CSV iterator for each file
		csvfile, err := csvparser.NewCSViter(fmt.Sprintf(dirname_ready + record))
		if err != nil {
			Util.Logger.Error("Trouble opening CSV file and or Iterator",
				"Location", "app.go - Range csvArray loop",
				"Record", record,
				"Error", err)
		}
		defer csvfile.Close()

		// At this point we have a csvfile iterator that can be used to read the CSV file line by line.
		for {
			// Start timer for begining of scraping
			funcDur := time.Now()

			// Read Row and index from CSVfile iterator
			row, indexer, _, err := csvfile.Next()

			//If error is End of File, break the loop. Go back to CSVarray loop
			if errors.Is(err, io.EOF) {
				Util.Logger.Error("Error Opening CSV/ CSV is empty - CONTINUE",
					slog.String("Location", "app.go - Range csvArray loop -> inside For loop"),
					slog.Any("Row", row),
					slog.Any("Error", err))
				break
				// if error is nil, continue with the row could be a empty row or an invalid row
			} else if err != nil {
				Util.Logger.Error("Error Opening CSV/ Reading Row - keeps running",
					slog.String("Location", "app.go - Range csvArray loop -> inside For loop"),
					slog.Any("Row", row),
					slog.Any("Error", err))
				continue
			}
			//Caching logic, find index and if it is not the cache file, skip the file
			if cacheFile != "" && record == cacheFile && cacheIndexInt >= 0 && indexer < cacheIndexInt {
				// still before the resume point → skip
				continue
			}
			if cacheFile != "" && record == cacheFile && cacheIndexInt == indexer {
				Util.Logger.Debug("Resuming at cached index",
					"Record", record, "Index", indexer, "CacheIndex", cacheIndexInt)
				// after this iteration, subsequent lines (> cacheIndexInt) will flow naturally
			}
			//End caching logic -------------------------------->

			//If starts with http its a url try and scrape with it.
			if len(row) == 0 || !strings.HasPrefix(row[0], "http") {
				continue
			}

			// this gets the first element of the row which is the URL
			url := row[0]

			// Caching URL and file name structure
			curCache := &Util.TrackCache{
				CurrentURL:  url,
				CurrentFile: record,
				Index:       indexer,
				Sport:       "Soccer",
				CacheType:   "current",
			}
			// Build a fresh context for THIS URL
			reqCtx := colly.NewContext()
			reqCtx.Put("Record", record)
			reqCtx.Put("Sport", curCache.Sport)
			reqCtx.Put("CacheType", curCache.CacheType)
			reqCtx.Put("Index", strconv.Itoa(indexer)) // Put takes string
			reqCtx.Put("Url", curCache.CurrentURL)
			reqCtx.Put("retryCount", 0)

			// Start Caching current links and URLS
			// Create a temporary directory, tmp/CollyClicker/Sport
			// dont crash if temp directory does not work, just log it and continue
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
			// --- Per-request scraping state + handlers on a CLONE to avoid stacking
			keeperCounter := 0
			pageData := []scraper.TeamData{{}, {}}

			pageC := c.Clone()

			// --------------- Start: Colly Collector context per URL set up----------------------------->
			//On request set additional headers
			pageC.OnRequest(func(r *colly.Request) {
				r.Headers.Set("Referer", "https://fbref.com/")
				r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
				// set context for each request to track retry count
				// .(int) is a type assertion to get the retry count
				if _, ok := r.Ctx.GetAny("retryCount").(int); !ok {
					r.Ctx.Put("retryCount", 0)
				}

			}) //End of onREqeust funciton

			pageC.OnError(func(r *colly.Response, err error) {
				//Get all context values

				if r == nil {
					// network or other error before response
					Util.Logger.Error("Request failed before response", "err", err)
					return
				}
				//Ctx is a per request context, place after r==nil check to avoid nil pointer dereference
				// Safe pulls from context (Ctx can be nil; keys may be missing)
				var record, url string
				if r.Ctx != nil {
					record = r.Ctx.Get("Record")
					url = r.Ctx.Get("Url")
				}
				if r.StatusCode != 429 {
					Util.Logger.Error("Colly request error",
						"url", url,
						"status", r.StatusCode,
						"err", err,
					)
					return
				}
				Util.Logger.Error("Colly request error",
					"url", url,
					"status", r.StatusCode,
					"err", err,
				)
				// r.ctx is a per request context, so we can store retry count
				// "retryCount" is a key in the context to track retries
				// .(int) is a type assertion to get the retry count
				// no need for global variable, as each request has its own context
				// no race condition, as each request has its own context
				retryCount, _ := r.Ctx.GetAny("retryCount").(int)
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

				// wait is calculated as sleep + backoff + jitter
				//429 errors are rate limites, backoff is obtained from the Retry-After header
				time.Sleep(wait)
				r.Ctx.Put("retryCount", retryCount+1)
				_ = r.Request.Retry()

			}) //end of on error handler

			// --------------- End of Colly Collector context per URL set up----------------------------->
			for _, h := range scraper.GetSelectorHandlers(&pageData, &keeperCounter, &fbref) {
				h := h // shadow to avoid loop-var capture
				pageC.OnHTML(h.Selector, func(e *colly.HTMLElement) { h.Handler(e) })
			}
			// write once the page is fully scraped
			dateStr, _ := fileutils.ExtractDateFromURL(url)
			pageC.OnScraped(func(r *colly.Response) {
				scraper.PageDataToCSV(pageData, dateStr)
				Util.Logger.Info("Wrote pageData", "rows", len(pageData), "url", r.Request.URL.String())
			})

			// Send the request with context and time it here (url is in scope)
			urlStart := time.Now()
			if err := pageC.Request("GET", url, nil, reqCtx, nil); err != nil {
				Util.Logger.Error("Request enqueue failed", "url", url, "file", record, "err", err)
				continue
			}
			Util.Logger.Info("Finished URL Visit", "URL", url, "Duration", time.Since(urlStart))
			t := time.Now()
			elapsed := t.Sub(start)
			Util.Logger.Info("Finished scraping Handler ,",
				"Total Duration", elapsed,
				"Scrape time", time.Since(funcDur),
				//"Handler", handlers)
			)
		} //End of URL for loop
	} //End of CSV array for loop

	// Start time, Log total duration of all URLS
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
