package scraper

import (
	colly "github.com/gocolly/colly/v2"
)

// ScraperConfig holds configuration for CollyScraper,
// including the Colly collector, proxy options, and a CSS selector.
type ScraperConfig struct {
	Collector     *colly.Collector  // A pointer to Colly's collector object (created via colly.NewCollector()).
	UseProxy      bool              // Whether or not to use a proxy.
	ProxyList     []string          // List of proxy addresses to rotate through.
	LinkSelectors []SelectorHandler // CSS selector for the target HTML element(s).
	Debug         bool              // Enables optional debug logging.
}

// CollyScraper encapsulates a configured Colly collector.
type CollyScraper struct {
	Config *ScraperConfig
}

// Input all HTML elements at once, defining the link handlers up front
// 1 Scrape multiple elements
type SelectorHandler struct {
	Selector string
	Handler  func(e *colly.HTMLElement)
}

// NewCollyScraper is a constructor that accepts a fully prepared ScraperConfig.
// Example usage:
// cfg := &ScraperConfig{Collector: colly.NewCollector(), ...}
// s := scraper.NewCollyScraper(cfg)
func NewCollyScraper(cfg *ScraperConfig) *CollyScraper {

	for _, sh := range cfg.LinkSelectors {
		cfg.Collector.OnHTML(sh.Selector, sh.Handler)
	}
	return &CollyScraper{
		Config: cfg,
	}
}

// Scrape visits the provided URL and applies the user supplied handler
// to each element matching the LinkSelector.
func (s *CollyScraper) Scrape(url string) error {
	err := s.Config.Collector.Visit(url)
	s.Config.Collector.Wait()

	return err
}
