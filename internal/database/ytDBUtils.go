package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func AddYtItemToHistory(item DatabaseYTItem) error {

	db, err := getDatabaseInstance()
	if err != nil {
		return errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT MAX(sid) FROM YtHistory")

	var maxSid int
	var minSid int

	err = row.Scan(&maxSid)

	if err != nil {
		maxSid = 0
	}

	row = db.QueryRow("SELECT MIN(sid) FROM YtHistory")
	err = row.Scan(&minSid)

	if err != nil {
		minSid = 0
	}

	maxSid += 1

	if maxSid-500 >= minSid {
		_, _ = db.Exec("DELETE FROM YtHistory WHERE sid < ?", (maxSid - 500))
	}

	readAt := time.Now().String()

	_, err = db.Exec(`INSERT INTO YtHistory VALUES(?,?,?,?,?,?,?)`,
		maxSid,
		item.Link,
		item.ThumbNail,
		item.Title,
		item.Channel,
		item.PubDate.String(),
		readAt,
	)

	if err != nil {
		return errors.New("Error inserting into History table")
	}

	return nil
}

func GetYTReadItemsAsJson(channel string) (string, error) {
	db, err := getDatabaseInstance()
	if err != nil {
		return "", errors.New("Error getting database instance")
	}

	var rows *sql.Rows
	if channel == "all" {
		rows, err = db.Query("SELECT * FROM YtHistory ORDER BY sid DESC")
	} else {
		rows, err = db.Query("SELECT * FROM YtHistory WHERE channel=? ORDER BY sid DESC", channel)
	}

	if err != nil {
		return "", errors.New("Error retrieving from history")
	}

	returnSlice := make([]DatabaseYTSeenItem, 0)
	var seenItem DatabaseYTSeenItem

	for rows.Next() {
		rows.Scan(&seenItem.ID, &seenItem.Link, &seenItem.ThumbNail, &seenItem.Title, &seenItem.Channel, &seenItem.PubDate, &seenItem.SeenAt)
		seenItem.IsFavourite, _ = isYTFavourite(seenItem.Link)
		returnSlice = append(returnSlice, seenItem)
	}

	content, err := json.Marshal(returnSlice)
	if err != nil {
		return "", errors.New("Error marshalling into JSON")
	}

	return string(content), nil
}

func IsYtItemInHistory(link string) (bool, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return false, errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT url FROM YtHistory WHERE url=?", link)

	var u string
	err = row.Scan(&u)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func AddToYtFavourites(link string) error {
	db, err := getDatabaseInstance()

	if err != nil {
		return errors.New("Error getting database instance")
	}

	var item DatabaseYTSeenItem

	row := db.QueryRow("SELECT * FROM YtHistory WHERE url=?", link)
	err = row.Scan(&item.ID, &item.Link, &item.ThumbNail, &item.Title, &item.Channel, &item.PubDate, &item.SeenAt)

	if err != nil {
		return errors.New("Error retrieving item from history table")
	}

	favouritedAt := time.Now().String()

	_, err = db.Exec(`INSERT INTO YtFavourites VALUES(?,?,?,?,?,?)`,
		item.Link,
		item.ThumbNail,
		item.Title,
		item.Channel,
		item.PubDate,
		favouritedAt,
	)

	if err != nil {
		return errors.New("Error inserting into favourites table")
	}

	return nil
}

func RemoveFromYtFavourites(link string) error {
	db, err := getDatabaseInstance()

	if err != nil {
		return errors.New("Error getting database instance")
	}

	_, err = db.Exec("DELETE FROM YtFavourites WHERE url=?", link)

	if err != nil {
		return errors.New("Error deleting from favourites table")
	}

	return nil
}

func GetYtFavourites(channel string) (string, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return "", errors.New("Error getting database instance")
	}

	var rows *sql.Rows

	if channel == "all" {
		rows, err = db.Query("SELECT * FROM YtFavourites")
	} else {
		rows, err = db.Query("SELECT * FROM YtFavourites WHERE channel=?", channel)
	}

	if err != nil {
		return "", errors.New("Error getting from favourites table")
	}

	var itemList []DatabaseYTFavourite
	var item DatabaseYTFavourite

	for rows.Next() {
		err = rows.Scan(&item.Link, &item.ThumbNail, &item.Title, &item.Channel, &item.PubDate, &item.FavouritedAt)
		itemList = append(itemList, item)
	}

	jsonContent, err := json.Marshal(itemList)
	if err != nil {
		return "", errors.New("Error marshalling favourites YT")
	}

	return string(jsonContent), nil
}

func isYTFavourite(link string) (bool, error) {
	db, err := getDatabaseInstance()

	if err != nil {
		return false, errors.New("Error getting database instance")
	}

	row := db.QueryRow("SELECT url FROM YtFavourites WHERE url=?", link)

	var u string
	err = row.Scan(&u)

	if err == nil {
		return true, nil
	}

	return false, nil
}
