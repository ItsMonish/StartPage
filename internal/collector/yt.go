package collector

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ItsMonish/StartPage/internal/database"
	"github.com/ItsMonish/StartPage/internal/types"
)

var (
	chList      []string
	totalFeed   []types.JsonYtItem
	channelFeed map[string][]types.JsonYtItem

	strTotalFeed string
	strChFeed    map[string]string
)

func InitYtCollector(logger *log.Logger, list []types.ConfigTitleURLItem) {
	strTotalFeed = "[]"
	strChFeed = make(map[string]string)
	channelFeed = make(map[string][]types.JsonYtItem)
	totalFeed = make([]types.JsonYtItem, 0)

	err := LoadYtFromCache()
	if err != nil {
		logger.Println("Error in loading YT feed from cache")
		logger.Println(err.Error())
	}
	LoadYtSources(list)
	RefreshYtFeed(logger, list)
}

func RefreshYtFeed(logger *log.Logger, list []types.ConfigTitleURLItem) {
	for _, channel := range list {
		logger.Println("Collecting YT feed from", channel.Title)

		content, err := MakeRequest(channel.Url)
		if err != nil {
			logger.Println("Error collecting feed from", channel.Title)
		}

		var xmlFeed types.XmlYtFeed
		err = xml.Unmarshal([]byte(content), &xmlFeed)
		if err != nil {
			logger.Println("Unmarshalling error from YT Feed")
			logger.Println(err.Error())
		}

		var jItem types.JsonYtItem
		for _, item := range xmlFeed.Feed {
			jItem.Title = item.Title
			jItem.Link = item.Link.Value
			jItem.Channel = channel.Title
			jItem.Thumbnail = item.Thumbnail.Value
			jItem.PubDate, _ = time.Parse(time.RFC3339, item.PubDate)

			totalFeed = append(totalFeed, jItem)
			channelFeed[channel.Title] = append(channelFeed[channel.Title], jItem)
		}
	}

	sort.SliceStable(totalFeed, func(i, j int) bool {
		return totalFeed[i].PubDate.After(totalFeed[j].PubDate)
	})

	err := marshalAndUpdateYtFeeds()
	if err != nil {
		logger.Println(err)
	}
}

func LoadYtSources(list []types.ConfigTitleURLItem) {
	for _, channel := range list {
		chList = append(chList, channel.Title)
	}
}

func LoadYtFromCache() error {
	cacheFeed, err := database.GetYtCachedItems()
	if err != nil {
		return err
	}

	for _, item := range cacheFeed {
		item.ID = curId
		curId += 1
		totalFeed = append(totalFeed, item)
		channelFeed[item.Channel] = append(channelFeed[item.Channel], item)
	}

	err = marshalAndUpdateYtFeeds()
	if err != nil {
		return err
	}
	return nil
}

func GetYtSources() string {
	jsonCont, _ := json.Marshal(chList)
	return string(jsonCont)
}

func GetYtFullFeed() string {
	return strTotalFeed
}

func GetYtChannelFeed(channel string) (string, error) {
	con, ok := strChFeed[channel]
	if !ok {
		return "[]", fmt.Errorf("Channel %s not found", channel)
	}
	return con, nil
}

func marshalAndUpdateYtFeeds() error {
	jsonCont, err := json.Marshal(totalFeed)
	if err != nil {
		return err
	}
	strTotalFeed = string(jsonCont)

	for _, channel := range chList {
		jsonCont, err := json.Marshal(channelFeed[channel])
		if err != nil {
			return err
		}
		strChFeed[channel] = string(jsonCont)
	}

	return nil
}
