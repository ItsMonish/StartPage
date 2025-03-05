package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DB_FILE = "config/database.db"

var dbInstance *sql.DB

type DatabaseFeedItem struct {
	ID       int
	Title    string
	Link     string
	PubDate  time.Time
	Source   string
	Category string
}

type DatabaseFeedReadItem struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Link     string `json:"link"`
	PubDate  string `json:"pubDate"`
	ReadAt   string `json:"readAt"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

func AddToHistory(rssItem DatabaseFeedItem) error {
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
		rows, err = db.Query("SELECT * FROM RssHistory")
		if err != nil {
			return "{}", nil
		}
	} else if source == "" {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=?", category)
		if err != nil {
			return "{}", nil
		}
	} else {
		rows, err = db.Query("SELECT * FROM RssHistory WHERE category=? AND source=?", category, source)
		if err != nil {
			return "{}", nil
		}
	}

	var readItem DatabaseFeedReadItem
	var returnList []DatabaseFeedReadItem

	for rows.Next() {
		rows.Scan(&readItem.ID, &readItem.Link, &readItem.Title, &readItem.Source, &readItem.Category, &readItem.PubDate, &readItem.ReadAt)
		returnList = append(returnList, readItem)
	}

	jsonContent, err := json.Marshal(returnList)

	if err != nil {
		return "", errors.New("Error marshalling into json")
	}

	return string(jsonContent), nil
}

func getDatabaseInstance() (*sql.DB, error) {

	if dbInstance != nil {
		return dbInstance, nil
	}

	err := checkOrCreateDatabase()

	if err != nil {
		return nil, errors.New("Error creating database")
	}

	dbInstance, err = sql.Open("sqlite3", DB_FILE)

	if err != nil {
		return nil, errors.New("Error opening database")
	}

	_, err = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS RssHistory(
            sid INTEGER,
            url TEXT PRIMARY KEY,
            title TEXT,
            source TEXT,
            category TEXT,
            pubDate TEXT,
            readAt TEXT
        );
    `)

	_, err = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS RssFavorites(
            url TEXT PRIMARY KEY,
            title TEXT,
            source TEXT,
            category TEXT,
            pubDate TEXT,
            readAt TEXT
        );
    `)

	return dbInstance, nil
}

func checkOrCreateDatabase() error {

	_, err := os.Stat(DB_FILE)
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(DB_FILE)

		if err != nil {
			return errors.New("Database creation error")
		}

		defer f.Close()
	}

	return nil

}

func CloseDBInstance() {
	if dbInstance != nil {
		dbInstance.Close()
	}
}
