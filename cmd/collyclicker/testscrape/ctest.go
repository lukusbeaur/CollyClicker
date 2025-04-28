// collyclicker/app/ctest.go

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"collyclicker/internal/scraper"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func main() {
	//i hate this but i need to do this for production.
	var datapointCounter = 0
	var formationCounter = 0
	var tablesCounter = 0
	//var teamnameCounter = 0

	type TeamTables struct {
		Title   string
		TabName string
		Headers []string
		Rows    [][]string
	}
	type TeamData struct {
		Teamname     string
		CoachNames   []string
		CaptainNames []string
		Formation    string
		AllTables    [][]TeamTables
	}
	// Pre-allocate for two teams (Home and Away)
	var pageData = []TeamData{{}, {}}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	c.Limit((&colly.LimitRule{
		RandomDelay: 2 * time.Second,
		DomainGlob:  "*",
	}))

	//Specify the Selector: The element of the page you want
	//Specify the Handler: How do you want the data managed
	sh := []scraper.SelectorHandler{
		{
			//@ div.datapoint --> Coach and Captain
			Selector: "div.datapoint",
			Handler: func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)
				if strings.HasPrefix(text, "Manager:") {
					pageData[datapointCounter/2].CoachNames = append(pageData[datapointCounter/2].CoachNames, strings.TrimPrefix(text, "Manager:"))
				} else if strings.HasPrefix(text, "Captain:") {
					pageData[datapointCounter/2].CaptainNames = append(pageData[datapointCounter/2].CaptainNames, strings.TrimPrefix(text, "Captain:"))
				}
				datapointCounter++
			},
		},
		{
			//@ div.lineup th[colspan] --> Formation
			Selector: "div.lineup th[colspan]",
			Handler: func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)
				re := regexp.MustCompile(`\(([^\)]+)\)`)
				if text != "Bench" {
					match := re.FindStringSubmatch(text)
					if len(match) > 1 {
						pageData[formationCounter/2].Formation = match[1]
					}
				}
				formationCounter++
			},
		},
		{
			//@ div[id^='all_player_stats'] --> All Table Data
			Selector: "div[id^='all_player_stats']",
			Handler: func(e *colly.HTMLElement) {
				var Tables []TeamTables
				title := strings.TrimSpace(e.DOM.Find("h2").Eq(1).Text())
				// Save teamname into the TeamData

				e.DOM.Find("div.filter.switcher a.sr_preset").Each(func(_ int, tab *goquery.Selection) {
					var Table TeamTables
					Table.Title = title
					tabName := strings.TrimSpace(strings.ToLower(tab.Text()))
					suffix := strings.ReplaceAll(tabName, " ", "_")

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
					var matchedTable *goquery.Selection

					e.DOM.Find("table[id^='stats_']").Each(func(_ int, table *goquery.Selection) {
						id, exists := table.Attr("id")
						if !exists {
							return
						}
						if strings.Contains(id, suffix) {
							matchedTable = table
						}
					})
					if matchedTable == nil {
						fmt.Printf("No table found for tab %s\n", tabName)
						return
					}

					matchedTable.Find("thead tr").Each(func(_ int, header *goquery.Selection) {
						var headers []string
						if val, _ := header.Attr("class"); val == "over_header" {
							return
						}
						header.Find("th").Each(func(_ int, cell *goquery.Selection) {
							headers = append(headers, strings.TrimSpace(cell.Text()))
						})
						Table.Headers = headers
					})

					matchedTable.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
						var rowA []string
						row.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
							rowA = append(rowA, strings.TrimSpace(cell.Text()))
						})
						Table.Rows = append(Table.Rows, rowA)
					})
					Tables = append(Tables, Table)
				})

				pageData[tablesCounter/2].AllTables = append(pageData[tablesCounter/2].AllTables, Tables)
				tablesCounter++
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

	testlink := "https://fbref.com/en/matches/cc5b4244/Manchester-United-Fulham-August-16-2024-Premier-League"
	s := scraper.NewCollyScraper(cfg)

	err := s.Scrape(testlink)
	if err != nil {
		log.Printf("Error scraping %s : %v", testlink, err)
	}

	for teamIdx, team := range pageData {
		fmt.Printf("\n========== TEAM #%d ==========\n", teamIdx+1)
		fmt.Println("Team Name:", team.Teamname)
		fmt.Println("Coach Names:", strings.Join(team.CoachNames, ", "))
		fmt.Println("Captain Names:", strings.Join(team.CaptainNames, ", "))
		fmt.Println("Formation:", team.Formation)

		for tableGroupIdx, tableGroup := range team.AllTables {
			fmt.Printf("\n--- Table Group #%d ---\n", tableGroupIdx+1)
			for tableIdx, table := range tableGroup {
				fmt.Printf("\n   --- Table #%d ---\n", tableIdx+1)
				fmt.Println("   Title:", table.Title)
				fmt.Println("   Tab Name:", table.TabName)
				fmt.Println("   Headers:", strings.Join(table.Headers, ", "))
				fmt.Println("   Rows:")
				for _, row := range table.Rows {
					fmt.Println("    ", strings.Join(row, ", "))
				}
			}
		}
	}

}
