package collector

import (
	"io"
	"log"
	"net/http"

	"github.com/ItsMonish/StartPage/internal/types"
)

func StartCollectors(logger *log.Logger, conf types.RootConfiguration) {
	sources = make(map[string][]string)
	sourceFeed = make(map[string][]types.JsonFeedItem)
	catFeed = make(map[string][]types.JsonFeedItem)
	strSrcFeed = make(map[string]string)
	strCatFeed = make(map[string]string)
	jsonFeed = make([]types.JsonFeedItem, 0)

	strTotalFeed = "[]"
	strChFeed = make(map[string]string)
	channelFeed = make(map[string][]types.JsonYtItem)
	totalFeed = make([]types.JsonYtItem, 0)

	err := LoadRssFromCache()
	if err != nil {
		logger.Println("Error in loading RSS feed from cache")
		logger.Println(err.Error())
	}
	err = LoadYtFromCache()
	if err != nil {
		logger.Println("Error in loading YT feed from cache")
		logger.Println(err.Error())
	}

	LoadRssSources(conf.Rss)
	RefreshRssFeed(logger, conf.Rss)

	LoadYtSources(conf.Yt)
	RefreshYtFeed(logger, conf.Yt)
}

func MakeRequest(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}
