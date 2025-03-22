package server

import (
	"log"
	"time"

	"github.com/ItsMonish/StartPage/internal/collector"
	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/database"
)

func updateWithInterval(interval int) time.Time {
	return time.Now().Add(time.Duration(interval) * time.Minute)
}
func startServerRoutine(logger *log.Logger, stopRoutine chan bool, conf config.Configuration) {
	nextRefresh := updateWithInterval(conf.Props.RefreshInterval)

	eFlag := collector.RefreshRssFeed(logger, conf.Rss)
	if eFlag {
		logger.Println("There was some error in collecting RSS feed. Retrying in", conf.Props.RetryInterval, "minutes")
		nextRefresh = updateWithInterval(conf.Props.RefreshInterval)
	} else {
		logger.Println("Collected from RSS sources successfully")
	}
	eFlag = collector.RefreshYtFeed(logger, conf.Yt)
	if eFlag {
		logger.Println("There was some error in collecting YT feed. Retrying in", conf.Props.RetryInterval, "minutes")
		nextRefresh = updateWithInterval(conf.Props.RefreshInterval)
	} else {
		logger.Println("Collected from YT sources successfully")
	}

	for {
		select {
		case <-stopRoutine:
			logger.Println("Stopping server routine")
			return
		default:
			if time.Now().After(nextRefresh) {
				collector.RefreshRssFeed(logger, conf.Rss)
				logger.Println("Colleced from RSS sources")
				collector.RefreshYtFeed(logger, conf.Yt)
				logger.Println("Collected from YT sources")
				nextRefresh = updateWithInterval(conf.Props.RefreshInterval)
			} else {
				time.Sleep(time.Minute)
			}
		}
	}
}

func convertToDBItem(item collector.JsonFeedItem) database.DatabaseFeedItem {
	var dbItem database.DatabaseFeedItem
	dbItem.ID = item.ID
	dbItem.Title = item.Title
	dbItem.Link = item.Link
	dbItem.Source = item.Source
	dbItem.Category = item.Category
	dbItem.PubDate = item.PubDate

	return dbItem
}

func convertYTtoDBItem(item collector.JsonYtItem) database.DatabaseYTItem {
	var dbItem database.DatabaseYTItem
	dbItem.ID = item.ID
	dbItem.Title = item.Title
	dbItem.Link = item.Link
	dbItem.Channel = item.Channel
	dbItem.PubDate = item.PubDate
	dbItem.ThumbNail = item.ThumbNail

	return dbItem
}
