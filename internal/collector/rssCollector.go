package collector

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/ItsMonish/StartPage/internal/config"
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
	Title   string    `json:"title"`
	Link    string    `json:"link"`
	PubDate time.Time `json:"pubDate"`
	Source  string    `json:"source"`
}

func CollectRssFeed(logger *log.Logger, rssList map[string][]config.TitleURLItem) string {
	var xmlFeeds []XmlRssFeed

	for category, items := range rssList {
		for _, item := range items {
			resp, err := http.Get(item.Url)

			if err != nil {
				logger.Println("Error collecting from", item.Url)
			}

			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			var xmlFeed XmlRssFeed

			xmlFeed.Source = item.Title
			xmlFeed.Category = category

			if err := xml.Unmarshal(body, &xmlFeed); err != nil {
				logger.Println("Error unmarshalling contents from", item.Url)
			}

			xmlFeeds = append(xmlFeeds, xmlFeed)

		}
	}

	var jsonItems []JsonFeedItem

	for _, feedItem := range xmlFeeds {
		src := feedItem.Source

		for _, feed := range feedItem.Feed {
			var item JsonFeedItem

			item.Source = src
			item.Title = feed.Title
			item.Link = feed.Link

			var err error

			item.PubDate, err = time.Parse(time.RFC1123Z, feed.PubDate)

			if err != nil {
				item.PubDate, _ = time.Parse(time.RFC1123, feed.PubDate)
			}

			jsonItems = append(jsonItems, item)
		}
	}

	sort.SliceStable(jsonItems, func(i, j int) bool {
		return jsonItems[i].PubDate.After(jsonItems[j].PubDate)
	})

	jsonContent, err := json.Marshal(jsonItems)

	if err != nil {
		logger.Println("Error in Marshalling JSON for RSS")
	}

	return string(jsonContent)
}
