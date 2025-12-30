package database

import (
	"errors"
	"time"

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

func AddToRssHistory(rssItem types.JsonFeedItem) error {
	db, _ := GetDbInstance()
	row := db.QueryRow("SELECT MAX(sid) FROM RssHistory")

	var maxSid int
	err := row.Scan(&maxSid)
	if err != nil {
		maxSid = 0
	}

	var count int = 0
	row = db.QueryRow("SELECT COUNT(*) FROM RssHistory WHERE source=?", rssItem.Source)
	_ = row.Scan(&count)

	for count >= 75 {
		_, _ = db.Exec(`DELETE FROM RssHistory 
						WHERE 
						sid=(SELECT MIN(sid) FROM RssHistory 
							WHERE
							source=?)`, rssItem.Source)
		count--
	}

	maxSid += 1
	readAt := time.Now().String()

	_, err = db.Exec(`INSERT INTO RssHistory VALUES(?,?,?,?,?,?,?)`,
		maxSid,
		rssItem.Link,
		rssItem.Title,
		rssItem.Source,
		rssItem.Category,
		rssItem.PubDate.String(),
		readAt,
	)

	if err != nil {
		return errors.New("Error inserting into History table")
	}

	return nil
}
