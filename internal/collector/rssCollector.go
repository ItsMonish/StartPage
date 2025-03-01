package collector

import (
	"encoding/json"
	"encoding/xml"
	"errors"
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
	Title    string    `json:"title"`
	Link     string    `json:"link"`
	PubDate  time.Time `json:"pubDate"`
	Source   string    `json:"source"`
	Category string    `json:"category"`
}

var (
	rssJsonItems    []JsonFeedItem
	rssJsonString   string
	sources         map[string][]string
	sourcesAsString string

	sourceFeed map[string][]JsonFeedItem
)

func RefreshRssFeed(logger *log.Logger, rssList map[string][]config.TitleURLItem) {
	var xmlFeeds []XmlRssFeed

	if sources == nil {
		sources = make(map[string][]string)
	}
	if sourceFeed == nil {
		sourceFeed = make(map[string][]JsonFeedItem)
	}

	for category, items := range rssList {

		for _, item := range items {

			sources[category] = append(sources[category], item.Title)

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

	for _, feedItem := range xmlFeeds {
		src := feedItem.Source
		category := feedItem.Category

		for _, feed := range feedItem.Feed {
			var item JsonFeedItem

			item.Source = src
			item.Title = feed.Title
			item.Link = feed.Link
			item.Category = category

			var err error

			item.PubDate, err = time.Parse(time.RFC1123Z, feed.PubDate)

			if err != nil {
				item.PubDate, _ = time.Parse(time.RFC1123, feed.PubDate)
			}

			rssJsonItems = append(rssJsonItems, item)

			sourceFeed[src] = append(sourceFeed[src], item)
		}
	}

	sort.SliceStable(rssJsonItems, func(i, j int) bool {
		return rssJsonItems[i].PubDate.After(rssJsonItems[j].PubDate)
	})

	jsonContent, err := json.Marshal(rssJsonItems)

	if err != nil {
		logger.Println("Error in Marshalling JSON for RSS")
	}

	rssJsonString = string(jsonContent)

	if err != nil {
		logger.Println("Error in Marshalling JSON for Sources")
	}

	sourcesCont, err := json.Marshal(sources)

	sourcesAsString = string(sourcesCont)

	//for cat, srcs := range sources {
	//	logger.Println("category", cat)
	//	for _, src := range srcs {
	//		logger.Println("\t", src)
	//		for _, item := range sourceFeed[src] {
	//			logger.Println("\t\t", item.Link)
	//		}
	//	}
	//}

	//for cat, items := range sourceFeed {
	//	logger.Println("category", cat)
	//	for _, item := range items {
	//		logger.Println("item", item)
	//	}
	//}
}

func GetSourcesAsObj() map[string][]string {
	return sources
}

func GetSourcesAsStr() string {
	return sourcesAsString
}

func CollectRssAsJson() string {
	return rssJsonString
}

func CollectRssAsObj() []JsonFeedItem {
	return rssJsonItems
}

func GetSourceFeed(source string) (string, error) {
	jsonFeed, ok := sourceFeed[source]

	if !ok {
		return "", errors.New("not found")
	}

	jsonCont, _ := json.Marshal(jsonFeed)

	return string(jsonCont), nil

}

func GetCategoryFeed(category string) (string, error) {
	sources, ok := sources[category]

	if !ok {
		return "", errors.New("not found")
	}

	var categoryFeed []JsonFeedItem

	for _, source := range sources {
		for _, item := range sourceFeed[source] {
			categoryFeed = append(categoryFeed, item)
		}
	}

	sort.SliceStable(categoryFeed, func(i, j int) bool {
		return categoryFeed[i].PubDate.After(categoryFeed[j].PubDate)
	})

	jsonCont, _ := json.Marshal(categoryFeed)

	return string(jsonCont), nil
}
