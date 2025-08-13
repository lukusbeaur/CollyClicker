package scraper

import (
	"collyclicker/internal/fileutils"
	"fmt"
	"log"
	"regexp"
	"strings"

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

func GetSelectorHandlers(pageData *[]TeamData, keeperCounter *int, fbref *[]string) []SelectorHandler {
	return []SelectorHandler{
		{
			Name:     "Coach and Captian",
			Selector: "div.datapoint",
			Handler:  CapManHandler(pageData),
		},
		{
			Name:     "Line up",
			Selector: "div.lineup th[colspan]",
			Handler:  lineupHandler(pageData),
		},
		{
			Name:     "Player Stats",
			Selector: "div[id^='all_player_stats']",
			Handler:  playerStatsHandler(pageData),
		},
		{
			Name:     "Keeper Stats",
			Selector: "div[id^='all_keeper_stats_']",
			Handler:  keeperStatsHandler(pageData, keeperCounter),
		},
	}
}

func CapManHandler(pageData *[]TeamData) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if strings.HasPrefix(text, "Manager:") {
			if (*pageData)[0].CoachNames == "" {
				(*pageData)[0].CoachNames = strings.TrimPrefix(text, "Manager:")
			} else {
				(*pageData)[1].CoachNames = strings.TrimPrefix(text, "Manager:")
			}
		} else if strings.HasPrefix(text, "Captain:") {
			if (*pageData)[0].CaptainNames == "" {
				(*pageData)[0].CaptainNames = strings.TrimPrefix(text, "Captain:")
			} else {
				(*pageData)[1].CaptainNames = strings.TrimPrefix(text, "Captain:")
			}
		}
	}
}

func lineupHandler(pageData *[]TeamData) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		re := regexp.MustCompile(`\(([^\)]+)\)`)
		if text != "Bench" {
			match := re.FindStringSubmatch(text)
			if len(match) > 1 {
				if (*pageData)[0].Formation == "" {
					(*pageData)[0].Formation = match[1]
				} else {
					(*pageData)[1].Formation = match[1]
				}
			}
		}
	}
}

func playerStatsHandler(pageData *[]TeamData) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		var Tables []TeamTables
		title := strings.TrimSpace(e.DOM.Find("h2").Eq(1).Text())
		Teamname := strings.TrimSpace(e.DOM.Find("caption").Eq(1).Text())
		e.DOM.Find("div.filter.switcher a.sr_preset").Each(func(_ int, tab *goquery.Selection) {
			var Table TeamTables
			Table.Title = title
			Table.teamname = Teamname
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
				//fmt.Printf("No table found for tab %s\n", tabName)
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
		if (*pageData)[0].Teamname == "" {
			(*pageData)[0].Teamname = sanatizeTitle(Teamname)
			if len((*pageData)[0].AllTables) == 0 {
				(*pageData)[0].AllTables = append((*pageData)[0].AllTables, Tables)
			}

		} else if (*pageData)[0].Teamname == sanatizeTitle(Teamname) {
			if len((*pageData)[0].AllTables) == 0 {
				(*pageData)[0].AllTables = append((*pageData)[0].AllTables, Tables)
			}
		} else {
			(*pageData)[1].Teamname = sanatizeTitle(Teamname)
			if len((*pageData)[1].AllTables) == 0 {
				(*pageData)[1].AllTables = append((*pageData)[1].AllTables, Tables)
			}

		}
	}
}

func keeperStatsHandler(pageData *[]TeamData, keeperCounter *int) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		var Table TeamTables
		fullTitle := strings.TrimSpace(e.DOM.Find("h2").Text())
		Table.Title = fullTitle
		Table.TabName = "keeper_stats"
		teamName := sanatizeTitle(fullTitle)
		if (*pageData)[*keeperCounter].Teamname == "" {
			(*pageData)[*keeperCounter].Teamname = teamName
		}
		table := e.DOM.Find("table.stats_table")
		table.Find("thead tr").Each(func(_ int, tr *goquery.Selection) {
			if class, _ := tr.Attr("class"); class == "over_header" {
				return
			}
			tr.Find("th").Each(func(_ int, th *goquery.Selection) {
				Table.Headers = append(Table.Headers, strings.TrimSpace(th.Text()))
			})
		})
		table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			var rowData []string
			row.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
				rowData = append(rowData, strings.TrimSpace(cell.Text()))
			})
			Table.Rows = append(Table.Rows, rowData)
		})
		(*pageData)[*keeperCounter].AllTables = append((*pageData)[*keeperCounter].AllTables, []TeamTables{Table})
		*keeperCounter++
	}
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
