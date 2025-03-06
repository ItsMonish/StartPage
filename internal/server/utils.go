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

	collector.RefreshRssFeed(logger, conf.Rss)
	logger.Println("Collecting from RSS sources")

	for {
		select {
		case <-stopRoutine:
			logger.Println("Stopping server routine")
			return
		default:
			if time.Now().After(nextRefresh) {
				logger.Println("Collecting from RSS sources")
				collector.RefreshRssFeed(logger, conf.Rss)

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
