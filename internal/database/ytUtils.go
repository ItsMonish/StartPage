package database

import (
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
