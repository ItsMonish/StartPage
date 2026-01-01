package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ItsMonish/StartPage/internal/types"
)

func GetYtCachedItems() ([]types.JsonYtItem, error) {
	db, _ := GetDbInstance()

	test := db.QueryRow(`SELECT count(*) FROM YtCache`)
	var tmp string
	err := test.Scan(&tmp)
	if err != nil {
		return nil, nil
	}

	rows, err := db.Query(`SELECT * FROM YtCache`)
	if err != nil {
		return nil, err
	}

	var cacheItems []types.JsonYtItem
	var item types.JsonYtItem
	var timeHolder string

	for rows.Next() {
		err = rows.Scan(&item.Link, &item.Thumbnail, &item.Title, &item.Channel, &timeHolder)
		item.PubDate, _ = time.Parse(time.RFC3339, timeHolder)
		if err != nil {
			return nil, err
		}
		cacheItems = append(cacheItems, item)
	}

	return cacheItems, nil
}

func IsInYtHistory(url string) bool {
	db, _ := GetDbInstance()

	row := db.QueryRow(`SELECT sid FROM YtHistory WHERE url=?`, url)
	var temp string
	err := row.Scan(&temp)

	if err != nil {
		return false
	}
	return true
}

func IsInYtCache(url string) bool {
	db, _ := GetDbInstance()

	row := db.QueryRow(`SELECT title FROM YtCache WHERE url=?`, url)
	var temp string
	err := row.Scan(&temp)

	if err != nil {
		return false
	}
	return true
}

func AddToYtHistory(item types.JsonYtItem) error {
	db, _ := GetDbInstance()
	row := db.QueryRow("SELECT MAX(sid) FROM YtHistory")

	var maxSid int
	err := row.Scan(&maxSid)
	if err != nil {
		maxSid = 0
	}

	var count int = 0
	row = db.QueryRow("SELECT COUNT(*) FROM YtHistory WHERE channel=?", item.Channel)
	_ = row.Scan(&count)

	for count >= 75 {
		_, _ = db.Exec(`DELETE FROM YtHistory 
						WHERE 
						sid=(SELECT MIN(sid) FROM YtHistory 
							WHERE
							source=?)`, item.Channel)
		count--
	}

	maxSid += 1
	seenAt := time.Now().String()

	_, err = db.Exec(`INSERT INTO YtHistory VALUES(?,?,?,?,?,?,?,?,?)`,
		maxSid,
		item.Link,
		item.Thumbnail,
		item.Title,
		item.Channel,
		item.PubDate.String(),
		seenAt,
		0,
		"",
	)

	if err != nil {
		return err
	}

	_, _ = db.Exec(`DELETE FROM YtCache WHERE url=?`, item.Link)

	return nil
}

func GetYtSeenItems(channel string) string {
	db, err := GetDbInstance()

	var rows *sql.Rows
	if channel == "all" {
		rows, err = db.Query("SELECT * FROM YtHistory ORDER BY sid DESC")
		if err != nil {
			return "{}"
		}
	} else {
		rows, err = db.Query("SELECT * FROM YtHistory WHERE channel=? ORDER BY sid DESC", channel)
		if err != nil {
			return "{}"
		}
	}

	var seenItem types.DatabaseYtItem
	var resList []types.DatabaseYtItem

	for rows.Next() {
		rows.Scan(&seenItem.ID,
			&seenItem.Link,
			&seenItem.Thumbnail,
			&seenItem.Title,
			&seenItem.Channel,
			&seenItem.PubDate,
			&seenItem.SeenAt,
			&seenItem.IsFavourite,
			&seenItem.FavouritedAt,
		)
		resList = append(resList, seenItem)
	}

	jsonContent, _ := json.Marshal(resList)
	return string(jsonContent)
}

func FavouriteYtItem(url string) error {
	timeNow := time.Now().String()
	db, _ := GetDbInstance()

	_, err := db.Exec(`UPDATE YtHistory SET isFavourite=?, favouritedAt=? WHERE url=?`, true, timeNow, url)
	if err != nil {
		return err
	}

	return nil
}

func UnFavouriteYtItem(url string) error {
	db, _ := GetDbInstance()
	_, err := db.Exec(`UPDATE YtHistory SET isFavourite=?, favouritedAt=? WHERE url=?`, false, "", url)
	if err != nil {
		return err
	}

	return nil
}

func GetFavouritedYtItems(channel string) (string, error) {
	db, _ := GetDbInstance()

	var rows *sql.Rows
	var err error
	if channel == "all" {
		rows, err = db.Query(`SELECT * FROM YtHistory WHERE isFavourite=1`)
	} else {
		rows, err = db.Query(`SELECT * FROM YtHistory WHERE isFavourite=1 AND channel=?`, channel)
	}
	if err != nil {
		return "", err
	}

	var favItems []types.DatabaseYtItem
	var item types.DatabaseYtItem
	for rows.Next() {
		rows.Scan(&item.ID,
			&item.Link,
			&item.Thumbnail,
			&item.Title,
			&item.Channel,
			&item.PubDate,
			&item.SeenAt,
			&item.IsFavourite,
			&item.FavouritedAt,
		)
		favItems = append(favItems, item)
	}

	jsonCont, _ := json.Marshal(favItems)
	return string(jsonCont), nil
}

func WriteYtItemsToCache(ytItems []types.JsonYtItem) error {
	db, _ := GetDbInstance()

	_, _ = db.Exec(`DELETE FROM YtCache`)

	for _, item := range ytItems {
		_, err := db.Exec(`INSERT INTO YtCache VALUES(?,?,?,?,?)`,
			item.Link,
			item.Thumbnail,
			item.Title,
			item.Channel,
			item.PubDate,
		)

		if err != nil {
			return err
		}
	}
	return nil
}
