package collector

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"slices"
	"sort"
	"time"

	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/database"
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

var (
	rssJsonItems    []JsonFeedItem
	rssJsonString   string
	sources         map[string][]string
	sourcesAsString string
	sourceFeed      map[string][]JsonFeedItem
	CurrentId       int = 1
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
			sourceFeed[item.Title] = make([]JsonFeedItem, 0)

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
			item.ID = CurrentId

			var err error

			item.PubDate, err = time.Parse(time.RFC1123Z, feed.PubDate)

			if err != nil {
				item.PubDate, _ = time.Parse(time.RFC1123, feed.PubDate)
			}

			if truth, err := database.IsItemInHistory(item.Link); !truth {
				rssJsonItems = append(rssJsonItems, item)
				sourceFeed[src] = append(sourceFeed[src], item)
				CurrentId += 1
			} else if err != nil {
				logger.Println(err)
			}
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

	// for cat, srcs := range sources {
	// 	logger.Println("category", cat)
	// 	for _, src := range srcs {
	// 		logger.Println("\t", src)
	// 		for _, item := range sourceFeed[src] {
	// 			logger.Println("\t\t", item.Link)
	// 		}
	// 	}
	// }
	//
	// for cat, items := range sourceFeed {
	// 	logger.Println("source", cat)
	// 	for _, item := range items {
	// 		logger.Println("item", item)
	// 	}
	// }
}

func GetSourcesAsObj() map[string][]string {
	return sources
}

func GetSourcesAsStr() string {
	return sourcesAsString
}

func CollectRssAsJson() string {
	if len(rssJsonItems) == 0 || rssJsonItems == nil {
		return "[]"
	}
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

	if len(jsonFeed) == 0 || jsonFeed == nil {
		return "[]", nil
	}

	jsonCont, _ := json.Marshal(jsonFeed)

	return string(jsonCont), nil

}

func GetCategoryFeed(category string) (string, error) {
	categoryFeed, err := GetCategorySlice(category)

	if err != nil {
		return "", err
	}

	if len(categoryFeed) == 0 || categoryFeed == nil {
		return "[]", nil
	}

	sort.SliceStable(categoryFeed, func(i, j int) bool {
		return categoryFeed[i].PubDate.After(categoryFeed[j].PubDate)
	})

	jsonCont, _ := json.Marshal(categoryFeed)

	return string(jsonCont), nil
}

func GetRSSItem(id int) (JsonFeedItem, error) {
	for _, item := range rssJsonItems {
		if item.ID == id {
			return item, nil
		}
	}
	return *new(JsonFeedItem), errors.New("Item not found")
}

func RemoveFromList(id int) error {
	idx := 0
	for ; idx < len(rssJsonItems); idx++ {
		if rssJsonItems[idx].ID == id {
			break
		}
	}
	if idx == len(rssJsonItems) {
		return errors.New("Item not found with id")
	}
	item := rssJsonItems[idx]
	rssJsonItems = slices.Delete(rssJsonItems, idx, idx+1)
	idx = -1
	for i, item := range sourceFeed[item.Source] {
		if item.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	sourceFeed[item.Source] = slices.Delete(sourceFeed[item.Source], idx, idx+1)
	newContent, _ := json.Marshal(rssJsonItems)
	rssJsonString = string(newContent)
	return nil
}

func GetCategorySlice(category string) ([]JsonFeedItem, error) {
	sources, ok := sources[category]

	if !ok {
		return nil, errors.New("Category not found")
	}

	var categoryFeed []JsonFeedItem

	for _, source := range sources {
		for _, item := range sourceFeed[source] {
			categoryFeed = append(categoryFeed, item)
		}
	}

	return categoryFeed, nil
}

func GetSourceSlice(source string) ([]JsonFeedItem, error) {
	returnList, ok := sourceFeed[source]

	if !ok {
		return nil, errors.New("Source not found")
	}

	return returnList, nil
}
