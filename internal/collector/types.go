package collector

import (
	"time"
)

type XmlRssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

type XmlRssFeed struct {
	Source   string
	Category string
	Feed     []XmlRssItem `xml:"channel>item"`
}

type JsonFeedItem struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Link     string    `json:"link"`
	PubDate  time.Time `json:"pubDate"`
	Source   string    `json:"source"`
	Category string    `json:"category"`
}
