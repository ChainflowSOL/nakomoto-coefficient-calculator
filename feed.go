package main

import (
	"encoding/xml"
	"fmt"
	"sort"
	"time"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	AtomLink      atomLink  `xml:"atom:link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Items         []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Category    string `xml:"category,omitempty"`
}

func renderRSS(baseURL string, coefficients []JsonResponse, updated time.Time) ([]byte, error) {
	sorted := make([]JsonResponse, len(coefficients))
	copy(sorted, coefficients)
	sort.Slice(sorted, func(i, j int) bool {
		ai, aj := abs(sorted[i].Change), abs(sorted[j].Change)
		if ai != aj {
			return ai > aj
		}
		return sorted[i].ChainToken < sorted[j].ChainToken
	})

	items := make([]rssItem, 0, len(sorted))
	pub := updated.UTC().Format(time.RFC1123Z)
	dayStamp := updated.UTC().Format("2006-01-02")

	for _, c := range sorted {
		var title, desc string
		if c.Change != 0 && c.NakaCoPrevVal != 0 {
			arrow := "↑"
			if c.Change < 0 {
				arrow = "↓"
			}
			title = fmt.Sprintf("%s Nakamoto Coefficient: %d → %d (%s%d)",
				c.ChainName, c.NakaCoPrevVal, c.NakaCoCurrVal, arrow, abs(c.Change))
			desc = fmt.Sprintf("The Nakamoto coefficient for %s (%s) changed from %d to %d.",
				c.ChainName, c.ChainToken, c.NakaCoPrevVal, c.NakaCoCurrVal)
		} else {
			title = fmt.Sprintf("%s Nakamoto Coefficient: %d", c.ChainName, c.NakaCoCurrVal)
			desc = fmt.Sprintf("Current Nakamoto coefficient for %s (%s) is %d.",
				c.ChainName, c.ChainToken, c.NakaCoCurrVal)
		}

		items = append(items, rssItem{
			Title:       title,
			Link:        fmt.Sprintf("%s/?chain=%s", baseURL, c.ChainToken),
			GUID:        fmt.Sprintf("%s/feed/%s-%d-%s", baseURL, c.ChainToken, c.NakaCoCurrVal, dayStamp),
			PubDate:     pub,
			Description: desc,
			Category:    c.ChainToken,
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         "Nakaflow — Nakamoto Coefficients",
			Link:          baseURL,
			AtomLink:      atomLink{Href: baseURL + "/feed.xml", Rel: "self", Type: "application/rss+xml"},
			Description:   "Live Nakamoto coefficient values across major proof-of-stake blockchains.",
			Language:      "en-us",
			LastBuildDate: pub,
			Items:         items,
		},
	}

	body, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
