package types

import "time"

type XmlFeed struct {
	Channel struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type XmlAtomFeed struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Value string `xml:"href,attr"`
		} `xml:"link"`
		PubDate string `xml:"published"`
	} `xml:"entry"`
}

type JsonFeedItem struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Link     string    `json:"link"`
	PubDate  time.Time `json:"pubDate"`
	Source   string    `json:"source"`
	Category string    `json:"category"`
}
