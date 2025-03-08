package database

import (
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
