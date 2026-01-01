package collector

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"slices"
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

	strSources  string = ""
	strJsonFeed string = ""
	strSrcFeed  map[string]string
	strCatFeed  map[string]string

	RssErrFlag bool = false
)

func RefreshRssFeed(logger *log.Logger, list map[string][]types.ConfigTitleURLItem) {
	for category, feedList := range list {
		for _, item := range feedList {
			logger.Println("Collecting RSS feed from", item.Title)

			content, err := MakeRequest(item.Url)
			if err != nil {
				RssErrFlag = true
				logger.Println("Error in collecting feed from", item.Title)
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

	err := database.WriteRssItemsToCache(jsonFeed)
	if err != nil {
		logger.Println("Error writing cache items to database")
		logger.Println(err)
	}

	err = marshalAndUpdateRssFeeds()
	if err != nil {
		logger.Println(err)
	}
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
		catFeed[item.Category] = append(catFeed[item.Category], item)
	}

	err = marshalAndUpdateRssFeeds()
	if err != nil {
		return err
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

func GetRssSources() string {
	return strSources
}

func GetRssFullFeed() string {
	return strJsonFeed
}

func GetCategoryFeed(category string) (string, error) {
	if _, ok := strCatFeed[category]; !ok {
		return "", errors.New("Invalid Category")
	}
	return strCatFeed[category], nil
}

func GetSourceFeed(source string) (string, error) {
	if _, ok := strSrcFeed[source]; !ok {
		return "", errors.New("Invalid Source")
	}
	return strSrcFeed[source], nil
}

func GetFeedItemWithId(id int) (types.JsonFeedItem, error) {
	for _, item := range jsonFeed {
		if item.ID == id {
			return item, nil
		}
	}
	return *new(types.JsonFeedItem), fmt.Errorf("Item with ID %d not found", id)
}

func RemoveFeedItemWithId(id int) error {
	idx := 0
	for ; idx < len(jsonFeed); idx++ {
		if jsonFeed[idx].ID == id {
			break
		}
	}
	if idx == len(jsonFeed) {
		return errors.New("Item not found with id")
	}
	item := jsonFeed[idx]

	jsonFeed = slices.Delete(jsonFeed, idx, idx+1)
	newContent, _ := json.Marshal(jsonFeed)
	strJsonFeed = string(newContent)

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
	newContent, _ = json.Marshal(sourceFeed[item.Source])
	strSrcFeed[item.Source] = string(newContent)

	idx = -1
	for i, item := range sourceFeed[item.Category] {
		if item.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	catFeed[item.Category] = slices.Delete(catFeed[item.Category], idx, idx+1)
	newContent, _ = json.Marshal(catFeed[item.Category])
	strCatFeed[item.Category] = string(newContent)

	return nil
}

func GetAndRemoveRssItems(category string, source string) ([]types.JsonFeedItem, error) {
	var returnFeed []types.JsonFeedItem
	if category == "" && source == "" {
		returnFeed = make([]types.JsonFeedItem, len(jsonFeed))
		copy(returnFeed, jsonFeed)

		jsonFeed = slices.Delete(jsonFeed, 0, len(jsonFeed))
		strJsonFeed = "[]"
		for cat := range catFeed {
			catFeed[cat] = slices.Delete(catFeed[cat], 0, len(catFeed[cat]))
			strCatFeed[cat] = "[]"
		}
		for src := range sourceFeed {
			sourceFeed[src] = slices.Delete(sourceFeed[src], 0, len(sourceFeed[src]))
			strSrcFeed[src] = "[]"
		}

		return returnFeed, nil
	} else if category != "" && source == "" {
		if _, ok := catFeed[category]; !ok {
			return nil, fmt.Errorf("Cannot find RSS category %s", category)
		}
		returnFeed = make([]types.JsonFeedItem, len(catFeed[category]))
		copy(returnFeed, catFeed[category])

		catFeed[category] = slices.Delete(catFeed[category], 0, len(catFeed[category]))
		strCatFeed[category] = "[]"

		for _, item := range returnFeed {
			idx := -1
			for i, jItem := range jsonFeed {
				if item.ID == jItem.ID {
					idx = i
					break
				}
			}
			if idx != -1 {
				jsonFeed = slices.Delete(jsonFeed, idx, idx+1)
			}
		}
		jsonCont, _ := json.Marshal(jsonFeed)
		strJsonFeed = string(jsonCont)

		for _, src := range sources[category] {
			sourceFeed[src] = slices.Delete(sourceFeed[src], 0, len(sourceFeed[src]))
			strSrcFeed[src] = "[]"
		}

		return returnFeed, nil
	} else {
		if _, ok := sourceFeed[source]; !ok {
			return nil, fmt.Errorf("Cannot find RSS source %s", source)
		}
		returnFeed = make([]types.JsonFeedItem, len(sourceFeed[source]))
		copy(returnFeed, sourceFeed[source])

		sourceFeed[source] = slices.Delete(sourceFeed[source], 0, len((sourceFeed[source])))
		strSrcFeed[source] = "[]"

		for _, item := range returnFeed {
			idx := -1
			for i, jItem := range jsonFeed {
				if item.ID == jItem.ID {
					idx = i
					break
				}
			}
			cIdx := -1
			for i, cItem := range catFeed[item.Category] {
				if item.ID == cItem.ID {
					cIdx = i
					break
				}
			}
			if idx != -1 {
				jsonFeed = slices.Delete(jsonFeed, idx, idx+1)
			}
			if cIdx != -1 {
				catFeed[category] = slices.Delete(catFeed[category], cIdx, cIdx+1)
			}
		}

		jsonCont, _ := json.Marshal(jsonFeed)
		strJsonFeed = string(jsonCont)

		jsonCont, _ = json.Marshal(catFeed[category])
		strCatFeed[category] = string(jsonCont)

		return returnFeed, nil
	}
}

func marshalAndUpdateRssFeeds() error {
	jsonContent, err := json.Marshal(jsonFeed)
	if err != nil {
		return err
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

	srcJsonContent, err := json.Marshal(sources)
	if err != nil {
		return err
	}

	strSources = string(srcJsonContent)
	return nil
}

func isAtomFeed(feed string) bool {
	if strings.Contains(feed, "<feed") {
		return true
	} else {
		return false
	}
}
