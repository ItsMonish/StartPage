package database

import (
	"github.com/ItsMonish/StartPage/internal/types"
)

func GetRssCachedItems() ([]types.JsonFeedItem, error) {
	db, err := GetDbInstance()
	if err != nil {
		return nil, err
	}

	test := db.QueryRow(`SELECT * FROM RssCache`)
	err = test.Scan(nil)
	if err == nil {
		return nil, nil
	}

	rows, err := db.Query(`SELECT * FROM RssCache`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cacheItems []types.JsonFeedItem
	var item types.JsonFeedItem

	for rows.Next() {
		err = rows.Scan(&item.Link, &item.Title, &item.Source, &item.Category, &item.PubDate)
		if err != nil {
			return nil, err
		}
		cacheItems = append(cacheItems, item)
	}

	return cacheItems, nil
}

func IsInRssHistory(url string) bool {
	db, _ := GetDbInstance()

	row := db.QueryRow(`SELECT * FROM RssHistory WHERE url=?`, url)
	err := row.Scan(nil)

	if err != nil {
		return false
	}
	return true
}

func IsInRssCache(url string) bool {
	db, _ := GetDbInstance()

	row := db.QueryRow(`SELECT * FROM RssCache WHERE url=?`, url)
	err := row.Scan(nil)

	if err != nil {
		return false
	}
	return true
}
