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

type XmlAtomFeed struct {
	Feed []XmlAtomItem `xml:"entry"`
}

type XmlAtomItem struct {
	Title   string          `xml:"title"`
	Link    XmlAtomItemLink `xml:"link"`
	PubDate string          `xml:"published"`
}

type XmlAtomItemLink struct {
	Value string `xml:"href,attr"`
}

type JsonFeedItem struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Link     string    `json:"link"`
	PubDate  time.Time `json:"pubDate"`
	Source   string    `json:"source"`
	Category string    `json:"category"`
}

type XmlYtFeed struct {
	Title string      `xml:"title"`
	Feed  []XmlYtItem `xml:"entry"`
}

type XmlYtItem struct {
	Title     string          `xml:"group>title"`
	Link      YtVideoLink     `xml:"link"`
	PubDate   string          `xml:"published"`
	Thumbnail YtThumbnailLink `xml:"group>thumbnail"`
}

type YtVideoLink struct {
	Value string `xml:"href,attr"`
}

type YtThumbnailLink struct {
	Value string `xml:"url,attr"`
}

type JsonYtItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	PubDate   time.Time `json:"pubDate"`
	Channel   string    `json:"channel"`
	ThumbNail string    `json:"thumbnail"`
}
