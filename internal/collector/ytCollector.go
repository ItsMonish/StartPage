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

var (
	channelList []string
	channelFeed map[string][]JsonYtItem
	totalFeed   []JsonYtItem
	strFeed     string
	chListStr   string
	ID          int = 1
)

func RefreshYtFeed(logger *log.Logger, ytItems []config.TitleURLItem) {
	channelList = make([]string, 0)
	channelFeed = make(map[string][]JsonYtItem)
	totalFeed = make([]JsonYtItem, 0)
	strFeed = ""
	chListStr = ""

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

func GetChannelFeed(channel string) (string, error) {
	feed, ok := channelFeed[channel]

	if !ok {
		return "", errors.New("Channel not found:" + channel)
	}

	jsonContent, err := json.Marshal(feed)

	if err != nil {
		return "", errors.New("Error marshalling channel feed into JSON")
	}

	return string(jsonContent), nil
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

		if truth, err := database.IsYtItemInHistory(currentJsonItem.Link); !truth {
			returnList = append(returnList, currentJsonItem)
			*id++
		} else if err != nil {
			return nil, err
		}
	}

	return returnList, nil
}

func GetYTItem(id int) (JsonYtItem, error) {
	for _, item := range totalFeed {
		if item.ID == id {
			return item, nil
		}
	}
	return *(new(JsonYtItem)), errors.New("Error getting item with ID")
}

func DeleteYTItem(id int) error {
	var idx int = -1

	for i, item := range totalFeed {
		if item.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.New("Error getting item with ID")
	}

	curItem := totalFeed[idx]

	totalFeed = slices.Delete(totalFeed, idx, idx+1)

	newContent, err := json.Marshal(totalFeed)
	if err != nil {
		return errors.New("Error Marshalling into JSON")
	}
	strFeed = string(newContent)

	idx = -1

	for i, item := range channelFeed[curItem.Channel] {
		if item.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.New("Error getting item with ID")
	}

	channelFeed[curItem.Channel] = slices.Delete(channelFeed[curItem.Channel], idx, idx+1)

	return nil
}

func GetYTFilterSlice(channel string) ([]JsonYtItem, error) {
	if channel == "all" {
		return totalFeed, nil
	}

	returnList, ok := channelFeed[channel]

	if !ok {
		return nil, errors.New("Error getting channel " + channel)
	}

	return returnList, nil
}
