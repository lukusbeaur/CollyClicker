// collyclicker/app/ctest.go

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"collyclicker/internal/fileutils"
	"collyclicker/internal/scraper"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type TeamTables struct {
	teamname string
	Title    string
	TabName  string
	Headers  []string
	Rows     [][]string
}
type TeamData struct {
	Teamname     string
	CoachNames   string
	CaptainNames string
	Formation    string
	AllTables    [][]TeamTables
}

func main() {
	var keeperCounter = 0
	// Pre-allocate for two teams (Home and Away)
	var pageData = []TeamData{{}, {}}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	c.Limit((&colly.LimitRule{
		RandomDelay: 10*time.Second + 5,
		DomainGlob:  "*",
	}))

	//Specify the Selector: The element of the page you want
	//Specify the Handler: How do you want the data managed
	sh := []scraper.SelectorHandler{
		{
			//@ div.datapoint(a div with a class name datapoint) --> captain / Manager names --> add them to struct
			Selector: "div.datapoint",
			Handler: func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)
				//Check to see if Selector has a prefix of Manager or captain,
				//if so --> append pageData Object slice @[handler specific counter /2] //Check logic notes above for why counter/2

				if strings.HasPrefix(text, "Manager:") {
					if pageData[0].CoachNames == "" {
						pageData[0].CoachNames = strings.TrimPrefix(text, "Manager:")
					} else {
						pageData[1].CoachNames = strings.TrimPrefix(text, "Manager:")
					}
				} else if strings.HasPrefix(text, "Captain:") {
					if pageData[0].CaptainNames == "" {
						pageData[0].CaptainNames = strings.TrimPrefix(text, "Captain:")
					} else {
						pageData[1].CaptainNames = strings.TrimPrefix(text, "Captain:")
					}
				}
			},
		},
		{
			//@ div.lineup (a div with a class name lineup) --> table header with the colspan attribute
			//--> Formation --> add them to struct.
			Selector: "div.lineup th[colspan]",
			Handler: func(e *colly.HTMLElement) {
				//Regex formation Manchester Utd (4-2-3-1) --> 4-2-3-1
				text := strings.TrimSpace(e.Text)
				re := regexp.MustCompile(`\(([^\)]+)\)`)
				//'Bench' is embedded in this Selector. != Bench assign Formation data.
				if text != "Bench" {
					match := re.FindStringSubmatch(text)
					if len(match) > 1 {
						if pageData[0].Formation == "" {
							pageData[0].Formation = match[1]
						} else {
							pageData[1].Formation = match[1]
						}
					}
				}
			},
		},
		{
			//@ div w/ ID all_player_stats --> select div w/ class filter & switcher --> the div with in that
			//-->table tabs --> add to TeamTables Object
			Selector: "div[id^='all_player_stats']",
			Handler: func(e *colly.HTMLElement) {
				var Tables []TeamTables
				// Save teamname into the TeamData
				title := strings.TrimSpace(e.DOM.Find("h2").Eq(1).Text())
				Teamname := strings.TrimSpace(e.DOM.Find("caption").Eq(1).Text())

				//@div--> Filter Class --> switcher Class -> Embedded a.sr_preset element
				//Find all tabs names.
				e.DOM.Find("div.filter.switcher a.sr_preset").Each(func(_ int, tab *goquery.Selection) {
					var Table TeamTables
					Table.Title = title
					Table.teamname = Teamname
					tabName := strings.TrimSpace(strings.ToLower(tab.Text()))
					suffix := strings.ReplaceAll(tabName, " ", "_")

					//So annoying but no nice way to do this.
					//The Table ID and Table Class names are not the same, as the tab class and tab ID names
					//So in order to make sure logic below works, We must manually change them. GROSS
					if suffix == "pass_types" {
						suffix = "passing_types"
					}
					if suffix == "defensive_actions" {
						suffix = "defense"
					}
					if suffix == "miscellaneous_stats" {
						suffix = "misc"
					}

					Table.TabName = suffix
					//A floating goQuery for a matched table.
					var matchedTable *goquery.Selection

					//@table /w ID starting with '_stats'
					e.DOM.Find("table[id^='stats_']").Each(func(_ int, table *goquery.Selection) {
						id, exists := table.Attr("id")
						if !exists {
							return
						}
						//if our table attribute 'id' contains the sanatized classname 'suffix',
						//assign matchedTable Query as table(our Each(*goquery))
						if strings.Contains(id, suffix) {
							matchedTable = table
						}
					})
					if matchedTable == nil {
						fmt.Printf("No table found for tab %s\n", tabName)
						return
					}
					//@matchedTable --> Find Table head --> Table row--> get all [th] table headers
					matchedTable.Find("thead tr").Each(func(_ int, header *goquery.Selection) {
						var headers []string
						//Skip the overheaders we dont need em
						if val, _ := header.Attr("class"); val == "over_header" {
							return
						}
						//
						header.Find("th").Each(func(_ int, cell *goquery.Selection) {
							headers = append(headers, strings.TrimSpace(cell.Text()))
						})
						Table.Headers = headers
					})
					//@matchedTable find the tbody --> Embeddeded Table rows --> th, td Cell data
					matchedTable.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
						var rowA []string
						row.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
							rowA = append(rowA, strings.TrimSpace(cell.Text()))
						})
						Table.Rows = append(Table.Rows, rowA)
					})
					//Tables slice will be appended with each table.
					Tables = append(Tables, Table)
				})
				if pageData[0].Teamname == "" {
					pageData[0].Teamname = sanatizeTitle(Teamname)
					pageData[0].AllTables = append(pageData[0].AllTables, Tables)
				} else if pageData[0].Teamname == sanatizeTitle(Teamname) {
					pageData[0].AllTables = append(pageData[0].AllTables, Tables)
				} else {
					pageData[1].Teamname = sanatizeTitle(Teamname)
					pageData[1].AllTables = append(pageData[1].AllTables, Tables)
				}
				Teamname = ""
			},
		},
		{
			Selector: "div[id^='all_keeper_stats_']",
			Handler: func(e *colly.HTMLElement) {
				var tables TeamTables

				fullTitle := strings.TrimSpace(e.DOM.Find("h2").Text())
				tables.Title = fullTitle
				tables.TabName = "keeper_stats"

				// Extract Team Name
				teamName := sanatizeTitle(fullTitle)
				if pageData[keeperCounter].Teamname == "" {
					pageData[keeperCounter].Teamname = teamName
				}

				// Find table
				table := e.DOM.Find("table.stats_table")

				// Headers
				table.Find("thead tr").Each(func(_ int, tr *goquery.Selection) {
					if class, _ := tr.Attr("class"); class == "over_header" {
						return
					}
					tr.Find("th").Each(func(_ int, th *goquery.Selection) {
						tables.Headers = append(tables.Headers, strings.TrimSpace(th.Text()))
					})
				})

				// Rows
				table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
					var rowData []string
					row.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
						rowData = append(rowData, strings.TrimSpace(cell.Text()))
					})
					tables.Rows = append(tables.Rows, rowData)
				})

				// Save into a new table group
				pageData[keeperCounter].AllTables = append(pageData[keeperCounter].AllTables, []TeamTables{tables})

				keeperCounter++ // move after saving
			},
		},
	}

	// Init constructor.
	cfg := &scraper.ScraperConfig{
		Collector:     c,
		UseProxy:      false,
		LinkSelectors: sh,
		Debug:         true,
	}

	s := scraper.NewCollyScraper(cfg)

	// Discover all CSV files inside /links/
	csvFiles, err := fileutils.Findcsvfiles("./links")
	if err != nil {
		log.Fatalf("Error discovering CSV files: %v", err)
	}

	for _, csvFile := range csvFiles {
		fullPath := "links/" + csvFile

		//Take CSV Folder[csvFiles] for each csvFile Found[csvFile]
		//ReadLinks takes all lines from CSV and returns the links-->[]urls
		links, err := fileutils.ReadLinksFromCSV(fullPath)
		if err != nil {
			log.Printf("Error reading links from %s: %v", csvFile, err)
			continue
		}
		//Each link inside links[]URL --> Get the data for the folder and file creation.
		for _, link := range links {

			dateStr, err := fileutils.ExtractDateFromURL(link)
			if err != nil {
				log.Printf("Error extracting date from link %s: %v", link, err)
				continue
			}

			//Start to scrape
			log.Printf("Scraping link: %s (Date: %s)", link, dateStr)

			// ---- SCRAPE ----
			err = s.Scrape(link)
			if err != nil {
				log.Printf("Error scraping link %s: %v", link, err)
				continue
			}
			PageDataToCSV(pageData, dateStr)
			keeperCounter = 0

		}
	}

	pageData = []TeamData{{}, {}}
}

func PageDataToCSV(pageData []TeamData, dateStr string) {
	outputDir := "output"

	for _, team := range pageData {
		if team.Teamname == "" {
			continue
		}
		teamFolder := fmt.Sprintf("%s/%s_%s", outputDir, sanitizeFolderName(team.Teamname), sanitizeFolderName(dateStr))

		infoHeaders := []string{"Teamname", "CoachNames", "CaptainNames", "Formation"}
		infoRow := []string{
			team.Teamname,
			team.CoachNames,
			team.CaptainNames,
			team.Formation,
		}
		infoRows := [][]string{infoRow}

		err := fileutils.WriteCSV(teamFolder, "team_info.csv", infoHeaders, infoRows)
		if err != nil {
			log.Printf("Failed to write team info for %s: %v", team.Teamname, err)
		}

		for _, tableGroup := range team.AllTables {
			for _, table := range tableGroup {
				fileName := fmt.Sprintf("%s_%s.csv",
					sanitizeFileName(team.Teamname),
					sanitizeFileName(table.TabName),
				)
				err := fileutils.WriteCSV(teamFolder, fileName, table.Headers, table.Rows)
				if err != nil {
					log.Printf("Failed to write table %s for %s: %v", table.TabName, team.Teamname, err)
				}
			}
		}
	}
}
func sanatizeTitle(word string) string {
	name := strings.Fields(word)
	if len(name) >= 2 {
		name = name[:len(name)-3]
	} else {
		log.Fatal("Array out of bounds ")
	}
	return strings.Join(name, " ")
}

// sanitizeFolderName makes a safe folder name (no spaces, etc.)
func sanitizeFolderName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

// sanitizeFileName makes a safe file name (removes spaces and special characters)
func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return name
}
