package database

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbInstance *sql.DB

func getDatabaseInstance() (*sql.DB, error) {

	if dbInstance != nil {
		return dbInstance, nil
	}

	DB_FILE, err := os.UserConfigDir()
	if err != nil {
		return nil, errors.New("User Config Folder not found")
	}
	DB_FILE = DB_FILE + "/startpage/database.db"

	err = checkOrCreateDatabase(DB_FILE)

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
        CREATE TABLE IF NOT EXISTS RssFavourites(
            url TEXT PRIMARY KEY,
            title TEXT,
            source TEXT,
            category TEXT,
            pubDate TEXT,
            favouritedAt TEXT
        );
    `)

	_, err = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS YtHistory(
			sid INTEGER,
            url TEXT PRIMARY KEY,
			thumbnail TEXT,
            title TEXT,
            channel TEXT,
            pubDate TEXT,
            seenAt TEXT
        );
    `)

	_, err = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS YtFavourites(
            url TEXT PRIMARY KEY,
			thumbnail TEXT,
            title TEXT,
            channel TEXT,
            pubDate TEXT,
            favouritedAt TEXT
        );
    `)

	return dbInstance, nil
}

func checkOrCreateDatabase(DB_FILE string) error {

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
