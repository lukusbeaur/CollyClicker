package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	u "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	//"github.com/gocolly/colly/debug"
)

func main() {
	starttime := time.Now()
	var totalsize int
	LogMessage("INFO", "Starting the scrapping process", nil)
	csvfiles, err := scanCSVFiles("../links/")
	if err != nil {
		log.Fatalf("Error scaning for CSV files: %s", err)
	}

	//Check last_URL for record != 'end' -> loop through main CSV file. Else Find index
	//of the last URL inside main CSV file. start loop from there. --------------------->
	file, reader := openFileReadAll("../last_URL.csv")
	defer file.Close()
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Unable to read the record form the file %s", err)
		return
	}
	if records[0][0] == "end" {
		//fmt.Println("Starting fresh....")
		LogMessage("INFO", "Starting from beginning of list", nil)
		for _, csvfilepath := range csvfiles {
			file, reader := openFileReadAll(csvfilepath)
			defer file.Close()
			records, err := reader.ReadAll()
			LogMessage("DEBUG", "Using a new file for links : "+csvfilepath, nil)
			if err != nil {
				log.Fatalf("Unable to read file %s:%s", csvfilepath, err)
			}
			//fmt.Printf("Working form %s", csvfilepath)
			for _, record := range records {
				url := record[0]
				//track current URL here
				writeCurrentURL(url, "../last_URL.csv", []string{file.Name()})
				//LogMessage("INFO", "Processing URL:"+url, nil)
				//fmt.Println("Processing URL:", url)
				totalsize += scrapeURL(url)
			}
		}
	} else {
		//find the index of record[0][0] The URL inside record[0][1] The file name and
		// path: This will find the index in which the loop will start with------------->
		//fmt.Printf("Cont: %s on url %s\n", records[1][0], records[0][0])
		LogMessage("INFO", "Continue on:"+records[1][0]+" on url"+records[0][0], nil)
		index, err := findLineIndex(records[1][0], records[0][0])
		if err != nil {
			LogMessage("ERROR", "There was an error with accessing the CSV", err)
			//log.Fatalf("There was an error with accessing the CSV %v", err)
			return
		}
		file, reader := openFileReadAll(records[1][0])
		LogMessage("INFO", "Processing File:"+records[1][0], nil)
		defer file.Close()
		records, err := reader.ReadAll()
		if err != nil {
			log.Fatalf("There was an error reading the csv data %v", err)
			return
		}
		for i := index; i < len(records); i++ {

			url := records[i][0]
			//track current URL here

			writeCurrentURL(url, "../last_URL.csv", []string{file.Name()})
			LogMessage("INFO", "Processing URL"+url, nil)
			//fmt.Println("Processing URL:", url)
			totalsize += scrapeURL(url)
		}
	}

	//fmt.Println("This if statement worked, it detected an 'end'")

	//Open tracker CSV file for URL. This will track where you are in the main list
	//and if there is an interuption the loop will start on the last saved URL -------->
	//Open Main CSV file for urls. Handle errors encapsulated in openfileandReadAll
	//Create a reader for the main CSV file, Save all records for iteration------------>
	//DONE, Checked. Remove on next push. create an if statement that if there is a URL inside the left off file then
	//cont: searchfor index of that url in the master link file and start loop from there.
	//Iterate through the records readall object. for each URL in the object scrape the
	//table and save it to the 'Datatype_teamname.csv ---------------------------------->
	//fmt.Printf("The time is %s. Starting to pull data\n", starttime)
	//TODO: create a graceful shutdown function. Channels, SIGS, and <- make.
	elapsedTime := time.Since((starttime))
	LogMessage("INFO", "Total Elapsed Time: "+fmt.Sprint(elapsedTime), nil)
	LogMessage("INFO", "Total File Size "+fmt.Sprint(totalsize), nil)
	//fmt.Printf("Total Elapsed time: %s\n", elapsedTime)
	//fmt.Printf("Total data collected: %d\n", totalsize)
}
func scrapeURL(url string) (totalsize int) {
	var teamName, dataType, season string
	LogMessage("INFO", "Start new URL", nil)
	//find team name and data type inside URL------------------------------------------>
	re := regexp.MustCompile(`/([a-z_]+)/([^/]+)-Match-Logs-`)
	dateRe := regexp.MustCompile(`\b(\d{4}(?:-\d{4})?)\b`)
	match := re.FindStringSubmatch(url)
	dmatch := dateRe.FindStringSubmatch(url)
	//if regex fails to pull team name or data type from url it is placed in an error
	//File. To keep track of potential erros. ------------------------------------------>
	if len(match) > 2 && len(dmatch) > 1 {
		teamName = match[2]
		dataType = match[1]
		season = dmatch[1]
	} else {
		teamName = "Teamname_error"
		dataType = "Datatype_error"
		season = "seasonDate_error"
		//dir := "../TeamData"
		appendToFile("../links/url_Failure.csv", []string{url})
		//fmt.Printf("Failed to extract data from URL: %s\n", url)
		LogMessage("ERROR", "Failed to extract all info from URL"+url, nil)
		return
	}
	LogMessage("INFO", "Extracted Data: "+dataType+"Extracted Team Name: "+teamName+"Extracted Season: "+season, nil)
	//fmt.Println("Extracted data type:", dataType)
	//fmt.Println("Extracted team name:", teamName)
	//fmt.Println("Extracted season Date:", season)

	//init the csv file writer and create files. Writer: for URL table data, fwriter to
	//cont: keep track of any errors on regex errors  ---------------------------------->
	dir := fmt.Sprintf("../../TeamData/%s/", season)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("unable to create directory %q: %s\n", dir, err)
	}
	_, writer := createFile(fmt.Sprintf("%s%s-%s.csv", dir, teamName, dataType))
	defer writer.Flush()

	//START: initiate a collector object
	//LogMessage("INFO", "Started new collector for fbref.com", nil)
	c := colly.NewCollector(
		colly.AllowedDomains("fbref.com"),
		colly.ParseHTTPErrorResponse(),
		//colly.Async(true),
		//colly.Debugger(&debug.LogDebugger{}),
	)
	//Trying to add a proxy and my internal network.
	//rp, err := proxy.RoundRobinProxySwitcher("192.168.1.203") //"http://185.133.250.195:8888",
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c.SetProxyFunc(rp)
	//LogMessage("INFO", "Proxies have been set.", nil)
	var startTime time.Time
	var requestSize int
	//lets try and connect first and print the call back of the request
	//Start timer
	proxies := []string{
		"http://185.133.250.195:8888",
		"",
	}
	c.WithTransport(&http.Transport{
		Proxy: func(req *http.Request) (*u.URL, error) {
			proxyStr := selectProxy(proxies)
			if proxyStr != "" {
				proxyURL, err := u.Parse(proxyStr)
				if err != nil {
					return nil, err
				}
				LogMessage("PROXY", "Using Proxy: "+proxyStr, nil)
				return proxyURL, nil
			}
			LogMessage("PROXY", "Using Proxy: Internal", nil)
			return nil, nil // Returning nil uses the internal network (no proxy)
		},
	})

	c.OnRequest(func(r *colly.Request) {
		startTime = time.Now()

		//fmt.Println("Visiting:", r.URL.String())
		//fmt.Printf("Proxy: %s\n", r.ProxyURL)
		LogMessage("INFO", "Visiting URL:"+r.URL.String(), nil)
		//LogMessage("INFO", "Proxy:"+r.ProxyURL, nil)
	})
	//on response lets check size of data
	c.OnResponse(func(r *colly.Response) {
		requestSize += len(r.Body)
		//fmt.Printf("Status code: %d\n", r.StatusCode)
		LogMessage("INFO", "Status Code:"+fmt.Sprint(r.StatusCode), nil)
		//If there is a status code other than 200 (OK) then add the URL to the error
		//list url_Failure.csv---------------------------------------------------------->
		if r.StatusCode != 200 {
			LogMessage("CAUTION", "Status code is not 200", nil)
			//fmt.Printf("Status code mustnt be 200 right: %d\n", r.StatusCode)
			LogMessage("CAUTION", "URL failed, Writing it to url_failed.csv  URL: "+url, nil)
			appendToFile("../links/url_Failure.csv", []string{url})
			//writer.Write([]string{url})
			//} else {
			//	fmt.Printf("Status code must be 200 right: %d", r.StatusCode)
		}
	})
	c.OnScraped(func(r *colly.Response) {
		elapsedTime := time.Since((startTime))
		LogMessage("INFO", "Scraped URL with "+fmt.Sprint(requestSize)+" in "+fmt.Sprint(elapsedTime), nil)
		//fmt.Printf("Elapsed Time for one URL: %s\n", elapsedTime)
		//fmt.Printf("Total Data pulled: %d bytes\n", requestSize)
	})

	//random delay rate limiter
	c.Limit(&colly.LimitRule{
		DomainGlob:  "fbref.com",
		Parallelism: 1,
		RandomDelay: 5*time.Second + 1,
	})

	//search for table and pull data, putting into CSV---------------------------------->
	c.OnHTML("#matchlogs_for", func(e *colly.HTMLElement) {
		//i need to keep track of the rows and columns for easy parcing.
		rowIndex := 0
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			var rowData []string
			parentTag := row.DOM.Parent().Nodes[0].Data

			if rowIndex == 0 {
				rowIndex++
				return
			}
			if row.Attr("class") == "spacer partial_table" {
				//fmt.Printf("Skipping row with class: %s\n", row.Attr("class"))
				return
			}
			if parentTag == "tfoot" {
				return
			}
			row.ForEach("th, td", func(_ int, cell *colly.HTMLElement) {
				rowData = append(rowData, cell.Text)

			})

			writeRecord(writer, rowData)
			rowIndex++
		})
	})
	//start and check for error
	err = c.Visit(url)
	if err != nil {
		LogMessage("ERROR", "Unable to visit site: "+url, err)
		appendToFile("../links/url_Failure.csv", []string{url})
		//fmt.Println("Error visiting the site:", err)

	}
	c.Wait()
	return requestSize
}
func createFile(filepath string) (*os.File, *csv.Writer) {
	file, err := os.Create(filepath)
	if err != nil {
		LogMessage("ERROR", "Unable to create file", err)
		log.Fatalf("Unable to create file %q: %s\n", filepath, err)
	}
	writer := csv.NewWriter(file)
	return file, writer
}
func writeRecord(writer *csv.Writer, record []string) {
	if err := writer.Write(record); err != nil {
		log.Fatalf("Unable to write record to file :%s\n", err)
	}
	writer.Flush()
}
func openFileReadAll(filepath string) (*os.File, *csv.Reader) {
	file, err := os.Open(filepath)
	if err != nil {
		LogMessage("ERROR", "Unable to locate file", err)
		log.Fatalf("Unable to open or locate file: %s\n", err)
	}
	reader := csv.NewReader(file)
	if err != nil {
		LogMessage("ERROR", "Unable to read from file", err)
		log.Fatalf("Unable to read file: %s", err)
	}
	return file, reader
}
func writeCurrentURL(record, filepath string, curfile []string) {
	//Create file, _ file object, returns writer object. this will overwrite ---------->
	file, writer := createFile(filepath)
	defer file.Close()
	defer writer.Flush()
	//write records to file' WriteRecord only works with arrays, so you musth convert-->
	records := []string{record}
	writeRecord(writer, records)
	writeRecord(writer, curfile)
}
func appendFileAndWriter(filePath string) (*os.File, *csv.Writer) {
	//This will open file and append lines instead of overwriting it ------------------>
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		LogMessage("ERROR", "Unable to open file during append", err)
		log.Fatalf("Unable to open file %q: %s\n", filePath, err)
	}
	writer := csv.NewWriter(file)
	return file, writer
}
func appendToFile(filepath string, record []string) {
	file, writer := appendFileAndWriter(filepath)
	defer file.Close()
	defer writer.Flush()

	writeRecord(writer, record)
}
func scanCSVFiles(dir string) ([]string, error) {
	var csvFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			LogMessage("ERROR", "Unable to scan CSV files", err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			csvFiles = append(csvFiles, path)
		}
		return nil
	})
	return csvFiles, err
}
func findLineIndex(filePath, value string) (int, error) {
	file, reader := openFileReadAll(filePath)
	defer file.Close()
	records, err := reader.ReadAll()
	if err != nil {
		LogMessage("ERROR", "Unable to read CSV Data", err)
		return -1, fmt.Errorf("unable to read CSV data: %v", err)
	}
	for i, record := range records {
		for _, field := range record {
			if field == value {
				return i, nil
			}
		}

	}
	LogMessage("ERROR", "Unable to find CSV", err)
	return -1, fmt.Errorf("Unable to find CSV")
}
func LogMessage(severity string, action string, err error) {
	//create, open, create writer, write record, defer close.
	logfile, _ := appendFileAndWriter("../app.log")
	defer logfile.Close()

	log.SetOutput(logfile)
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	logMessage := currentTime + " [" + severity + "] " + action

	if err != nil {
		logMessage += " - Error: " + err.Error()
	}
	// Write the log message to the file
	_, errWrite := logfile.WriteString(logMessage + "\n")
	if errWrite != nil {
		log.Fatalf("Failed to write log message %s", err)
	}
}

func selectProxy(proxies []string) string {
	rand.Seed(time.Now().UnixNano())
	return proxies[rand.Intn(len(proxies))]
}
