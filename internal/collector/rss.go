package collector

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/ItsMonish/StartPage/internal/database"
	"github.com/ItsMonish/StartPage/internal/types"
)

var (
	sources    map[string][]string
	catFeed    map[string][]types.JsonFeedItem
	sourceFeed map[string][]types.JsonFeedItem
	jsonFeed   []types.JsonFeedItem
	curId      int = 1

	strJsonFeed string = ""
	strSrcFeed  map[string]string
	strCatFeed  map[string]string
)

func InitRssCollector(logger *log.Logger, list map[string][]types.ConfigTitleURLItem) {
	sources = make(map[string][]string)
	sourceFeed = make(map[string][]types.JsonFeedItem)
	catFeed = make(map[string][]types.JsonFeedItem)
	strSrcFeed = make(map[string]string)
	strCatFeed = make(map[string]string)
	jsonFeed = make([]types.JsonFeedItem, 0)

	LoadRssSources(list)
	LoadRssFromCache()
	RefreshRssFeed(logger, list)
}

func RefreshRssFeed(logger *log.Logger, list map[string][]types.ConfigTitleURLItem) {
	for category, feedList := range list {
		for _, item := range feedList {
			logger.Println("Collecting RSS feed from ", item.Title)

			content, err := MakeRequest(item.Url)
			if err != nil {
				logger.Println("Error in collecting feed from ", item.Title)
				logger.Println(err.Error())
			}

			if isAtomFeed(content) {
				var atomFeed types.XmlAtomFeed
				if err = xml.Unmarshal([]byte(content), &atomFeed); err != nil {
					logger.Println("Error in unmarshalling XML from atom feed")
					logger.Println(err.Error())
				}

				for _, aItem := range atomFeed.Entries {
					if database.IsInRssCache(aItem.Link.Value) ||
						database.IsInRssHistory(aItem.Link.Value) {
						continue
					}
					var jsonItem types.JsonFeedItem
					jsonItem.Title = aItem.Title
					jsonItem.Link = aItem.Link.Value
					aTime, _ := time.Parse(time.RFC3339, aItem.PubDate)
					jsonItem.PubDate = aTime
					jsonItem.Source = item.Title
					jsonItem.Category = category
					jsonItem.ID = curId

					curId += 1
					jsonFeed = append(jsonFeed, jsonItem)
					catFeed[category] = append(catFeed[category], jsonItem)
					sourceFeed[item.Title] = append(sourceFeed[item.Title], jsonItem)
				}
			} else {
				var xmlFeed types.XmlFeed
				if err := xml.Unmarshal([]byte(content), &xmlFeed); err != nil {
					logger.Println("Error in unmarshalling XML from RSS feed")
					logger.Println(err.Error())
				}

				for _, xItem := range xmlFeed.Channel.Items {
					if database.IsInRssCache(xItem.Link) ||
						database.IsInRssHistory(xItem.Link) {
						continue
					}
					var jsonItem types.JsonFeedItem
					jsonItem.Title = xItem.Title
					jsonItem.Link = xItem.Link
					aTime, err := time.Parse(time.RFC1123, xItem.PubDate)
					if err != nil {
						aTime, _ = time.Parse(time.RFC1123Z, xItem.PubDate)
					}
					jsonItem.PubDate = aTime
					jsonItem.Source = item.Title
					jsonItem.Category = category
					jsonItem.ID = curId

					curId += 1
					catFeed[category] = append(catFeed[category], jsonItem)
					jsonFeed = append(jsonFeed, jsonItem)
					sourceFeed[item.Title] = append(sourceFeed[item.Title], jsonItem)
				}
			}
		}
	}

	sort.SliceStable(jsonFeed, func(i, j int) bool {
		return jsonFeed[i].PubDate.After(jsonFeed[j].PubDate)
	})

	jsonContent, err := json.Marshal(jsonFeed)
	if err != nil {
		logger.Println("Error in marshalling JSON feed")
		logger.Println(err.Error())
	}

	strJsonFeed = string(jsonContent)

	for source, feed := range sourceFeed {
		sort.SliceStable(feed, func(i, j int) bool {
			return feed[i].PubDate.After(feed[j].PubDate)
		})
		jSrcCont, _ := json.Marshal(feed)
		strSrcFeed[source] = string(jSrcCont)
	}

	for category, feed := range catFeed {
		sort.SliceStable(feed, func(i, j int) bool {
			return feed[i].PubDate.After(feed[j].PubDate)
		})
		jCatCont, _ := json.Marshal(feed)
		strCatFeed[category] = string(jCatCont)

	}

	logger.Println(strJsonFeed)
	logger.Println(strCatFeed)
	logger.Println(strSrcFeed)
}

func LoadRssFromCache() error {
	cacheFeed, err := database.GetRssCachedItems()
	if err != nil {
		return err
	}

	for _, item := range cacheFeed {
		item.ID = curId
		curId += 1
		jsonFeed = append(jsonFeed, item)
		sourceFeed[item.Source] = append(sourceFeed[item.Source], item)
	}
	return nil
}

func LoadRssSources(conf map[string][]types.ConfigTitleURLItem) {
	for category, urlItem := range conf {
		for _, item := range urlItem {
			if sources[category] == nil {
				sources[category] = make([]string, 0)
			}
			sources[category] = append(sources[category], item.Title)
		}
	}
}

func isAtomFeed(feed string) bool {
	if strings.Contains(feed, "<feed") {
		return true
	} else {
		return false
	}
}
