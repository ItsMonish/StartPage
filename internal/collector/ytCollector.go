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

var (
	channelList []string
	channelFeed map[string][]JsonYtItem
	totalFeed   []JsonYtItem
	strFeed     string
	chListStr   string
	ID          int = 1
)

func RefreshYtFeed(logger *log.Logger, ytItems []config.TitleURLItem) {
	if channelList == nil {
		channelList = make([]string, 0)
	}
	if channelFeed == nil {
		channelFeed = make(map[string][]JsonYtItem)
	}
	if totalFeed == nil {
		totalFeed = make([]JsonYtItem, 0)
	}

	for _, item := range ytItems {
		var fromXmlFeed XmlYtFeed

		channelList = append(channelList, item.Title)
		channelFeed[item.Title] = make([]JsonYtItem, 0)

		resp, err := http.Get(item.Url)

		if err != nil {
			logger.Println("Error getting youtube feed from", item.Title)
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if err := xml.Unmarshal(body, &fromXmlFeed); err != nil {
			logger.Println("Error unmarshalling contents from", item.Title)
			logger.Println(err)
		}

		jsonList, err := convertXMLtoJSONObject(fromXmlFeed.Feed, item.Title, &ID)

		if err != nil {
			logger.Println(err)
		}

		sort.SliceStable(jsonList, func(i, j int) bool {
			return jsonList[i].PubDate.After(jsonList[j].PubDate)
		})

		channelFeed[item.Title] = jsonList
		totalFeed = append(totalFeed, jsonList...)
	}

	sort.SliceStable(totalFeed, func(i, j int) bool {
		return totalFeed[i].PubDate.After(totalFeed[j].PubDate)
	})

	jsonContent, err := json.Marshal(totalFeed)

	if err != nil {
		logger.Println("Error marshalling contents into JSON")
	}

	strFeed = string(jsonContent)
}

func GetFullYtFeed() string {
	return strFeed
}

func GetYtChannelList() (string, error) {
	if chListStr != "" {
		return chListStr, nil
	}

	content, err := json.Marshal(channelList)

	if err != nil {
		return "", errors.New("Error marshalling channel list into JSON")
	}

	return string(content), nil
}

func convertXMLtoJSONObject(feed []XmlYtItem, channel string, id *int) ([]JsonYtItem, error) {
	returnList := make([]JsonYtItem, 0)

	var currentJsonItem JsonYtItem
	var err error

	for _, item := range feed {
		currentJsonItem.ID = *id
		currentJsonItem.Title = item.Title
		currentJsonItem.Link = item.Link.Value
		currentJsonItem.Channel = channel
		currentJsonItem.ThumbNail = item.Thumbnail.Value
		currentJsonItem.PubDate, err = time.Parse(time.RFC3339, item.PubDate)
		if err != nil {
			return nil, errors.New("Error parsing time")
		}

		returnList = append(returnList, currentJsonItem)
		*id++
	}

	return returnList, nil
}
