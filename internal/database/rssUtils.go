package database

import (
	"database/sql"
	"encoding/json"
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

	row := db.QueryRow(`SELECT sid FROM RssHistory WHERE url=?`, url)
	var temp string
	err := row.Scan(&temp)

	if err != nil {
		return false
	}
	return true
}

func IsInRssCache(url string) bool {
	db, _ := GetDbInstance()

	row := db.QueryRow(`SELECT title FROM RssCache WHERE url=?`, url)
	var temp string
	err := row.Scan(&temp)

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

	_, err = db.Exec(`INSERT INTO RssHistory VALUES(?,?,?,?,?,?,?,?,?)`,
		maxSid,
		rssItem.Link,
		rssItem.Title,
		rssItem.Source,
		rssItem.Category,
		rssItem.PubDate.String(),
		readAt,
		0,
		"",
	)

	if err != nil {
		return err
	}

	return nil
}

func GetRssViewed(category string, source string) string {
	db, err := GetDbInstance()

	var rows *sql.Rows
	if category == "" && source == "" {
		rows, err = db.Query("SELECT * FROM RssHistory ORDER BY sid DESC")
		if err != nil {
			return "{}"
		}
	} else if source == "" {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=? ORDER BY sid DESC", category)
		if err != nil {
			return "{}"
		}
	} else {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=? AND source=? ORDER BY sid DESC", category, source)
		if err != nil {
			return "{}"
		}
	}

	var readItem types.DatabaseRssItem
	var resList []types.DatabaseRssItem

	for rows.Next() {
		rows.Scan(&readItem.ID,
			&readItem.Link,
			&readItem.Title,
			&readItem.Source,
			&readItem.Category,
			&readItem.PubDate,
			&readItem.ReadAt,
			&readItem.IsFavourite,
			&readItem.FavouritedAt,
		)
		resList = append(resList, readItem)
	}

	jsonContent, _ := json.Marshal(resList)
	return string(jsonContent)
}

func FavouriteRssItem(link string) error {
	timeNow := time.Now().String()
	db, _ := GetDbInstance()

	_, err := db.Exec(`UPDATE RssHistory SET isFavourite=?, favouritedAt=? WHERE url=?`, true, timeNow, link)
	if err != nil {
		return err
	}

	return nil
}

func UnFavouriteRssItem(link string) error {
	db, _ := GetDbInstance()
	_, err := db.Exec(`UPDATE RssHistory SET isFavourite=?, favouritedAt=? WHERE url=?`, false, "", link)
	if err != nil {
		return err
	}

	return nil
}

func GetRssFavourites(category string, source string) (string, error) {
	db, _ := GetDbInstance()

	var rows *sql.Rows
	var err error
	if category == "all" && source == "" {
		rows, err = db.Query(`SELECT * FROM RssHistory WHERE isFavourite=1`)
	} else if category != "" && source == "" {
		rows, err = db.Query(`SELECT * FROM RssHistory WHERE isFavourite=1 AND category=?`, category)
	} else {
		rows, err = db.Query(`SELECT * FROM RssHistory WHERE isFavourite=1 AND source=?`, source)
	}
	if err != nil {
		return "", err
	}

	var favItems []types.DatabaseRssItem
	var item types.DatabaseRssItem
	for rows.Next() {
		rows.Scan(&item.ID,
			&item.Link,
			&item.Title,
			&item.Source,
			&item.Category,
			&item.PubDate,
			&item.ReadAt,
			&item.IsFavourite,
			&item.FavouritedAt,
		)
		favItems = append(favItems, item)
	}

	jsonCont, _ := json.Marshal(favItems)
	return string(jsonCont), nil
}
