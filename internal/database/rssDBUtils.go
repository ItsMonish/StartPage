package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func AddToHistory(rssItem DatabaseFeedItem) error {
	if rssItem.ID == 0 {
		return nil
	}

	db, err := getDatabaseInstance()

	if err != nil {
		return errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT MAX(sid) FROM RssHistory")

	var maxSid int
	var minSid int

	err = row.Scan(&maxSid)

	if err != nil {
		maxSid = 0
	}

	row = db.QueryRow("SELECT MIN(sid) FROM RssHistory")

	err = row.Scan(&minSid)

	if err != nil {
		minSid = 0
	}

	maxSid += 1

	if maxSid-500 >= minSid {
		_, _ = db.Exec("DELETE FROM RssHistory WHERE sid < ?", (maxSid - 500))
	}

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

func IsItemInHistory(link string) (bool, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return false, errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT url FROM RssHistory WHERE url=?", link)

	var u string
	err = row.Scan(&u)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func GetReadItemsAsJson(category string, source string) (string, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return "", errors.New("Error getting database instance")
	}

	var rows *sql.Rows
	if category == "" && source == "" {
		rows, err = db.Query("SELECT * FROM RssHistory ORDER BY sid DESC")
		if err != nil {
			return "{}", nil
		}
	} else if source == "" {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=? ORDER BY sid DESC", category)
		if err != nil {
			return "{}", nil
		}
	} else {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=? AND source=? ORDER BY sid DESC", category, source)
		if err != nil {
			return "{}", nil
		}
	}

	var readItem DatabaseFeedReadItem
	var returnList []DatabaseFeedReadItem

	for rows.Next() {
		rows.Scan(&readItem.ID, &readItem.Link, &readItem.Title, &readItem.Source, &readItem.Category, &readItem.PubDate, &readItem.ReadAt)
		readItem.IsFavourite, _ = isFavourite(readItem.Link)
		returnList = append(returnList, readItem)
	}

	jsonContent, err := json.Marshal(returnList)

	if err != nil {
		return "", errors.New("Error marshalling into json")
	}

	return string(jsonContent), nil
}

func AddToFavourites(link string) error {
	db, err := getDatabaseInstance()

	if err != nil {
		return errors.New("Error getting database instance")
	}

	var item DatabaseFeedReadItem

	row := db.QueryRow("SELECT * FROM RssHistory WHERE url=?", link)
	err = row.Scan(&item.ID, &item.Link, &item.Title, &item.Source, &item.Category, &item.PubDate, &item.ReadAt)

	if err != nil {
		return errors.New("Error retrieving item from history table")
	}

	favouritedAt := time.Now().String()

	_, err = db.Exec(`INSERT INTO RssFavourites VALUES(?,?,?,?,?,?)`,
		item.Link,
		item.Title,
		item.Source,
		item.Category,
		item.PubDate,
		favouritedAt,
	)

	if err != nil {
		return errors.New("Error inserting into favourites table")
	}

	return nil
}

func RemoveFromFavourites(link string) error {
	db, err := getDatabaseInstance()

	if err != nil {
		return errors.New("Error getting database instance")
	}

	_, err = db.Exec("DELETE FROM RssFavourites WHERE url=?", link)

	if err != nil {
		return errors.New("Error deleting from favourites table")
	}

	return nil
}

func GetFavourties(category string, source string) (string, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return "", errors.New("Error getting database instance")
	}

	var rows *sql.Rows

	if category == "all" && source == "" {
		rows, err = db.Query("SELECT * FROM RssFavourites")
	} else if source == "" {
		rows, err = db.Query("SELECT * FROM RssFavourites WHERE category=?", category)
	} else {
		rows, err = db.Query("SELECT * FROM RssFavourites WHERE category=? AND source=?", category, source)
	}

	if err != nil {
		return "", errors.New("Error getting from favourites table")
	}

	var itemList []FavouriteRSSItem
	var item FavouriteRSSItem

	for rows.Next() {
		err = rows.Scan(&item.Link, &item.Title, &item.Source, &item.Category, &item.PubDate, &item.FavouritedAt)
		itemList = append(itemList, item)
	}

	jsonContent, err := json.Marshal(itemList)
	if err != nil {
		return "", errors.New("Error marshalling favourites RSS")
	}

	return string(jsonContent), nil
}

func isFavourite(link string) (bool, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return false, errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT url FROM RssFavourites WHERE url=?", link)

	var u string
	err = row.Scan(&u)

	if err == nil {
		return true, nil
	}

	return false, nil
}
