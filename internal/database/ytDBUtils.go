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
		rows, err = db.Query("SELECT * FROM YtHistory")
	} else {
		rows, err = db.Query("SELECT * FROM YtHistory WHERE channel=?", channel)
	}

	if err != nil {
		return "", errors.New("Error retrieving from history")
	}

	returnSlice := make([]DatabaseYTSeenItem, 0)
	var seenItem DatabaseYTSeenItem

	for rows.Next() {
		rows.Scan(&seenItem.ID, &seenItem.Link, &seenItem.ThumbNail, &seenItem.Title, &seenItem.Channel, &seenItem.PubDate, &seenItem.SeenAt)
		returnSlice = append(returnSlice, seenItem)
		seenItem.IsFavourite, _ = isYTFavourite(seenItem.Link)
	}

	content, err := json.Marshal(returnSlice)
	if err != nil {
		return "", errors.New("Error marshalling into JSON")
	}

	return string(content), nil
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
