package database

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbInstance *sql.DB

func GetDbInstance() (*sql.DB, error) {
	if dbInstance != nil {
		return dbInstance, nil
	} else {
		return nil, errors.New("DB instance is not initialized")
	}
}

func InitDb(dbPath string) error {
	_, err := os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(dbPath)

		if err != nil {
			return err
		}
		defer f.Close()
	}

	dbInstance, err = sql.Open("sqlite3", dbPath)

	if err != nil {
		return err
	}

	createTablesIfNot()

	return nil
}

func createTablesIfNot() {
	_, _ = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS RssHistory(
            sid INTEGER,
            url TEXT PRIMARY KEY,
            title TEXT,
            source TEXT,
            category TEXT,
            pubDate TEXT,
            readAt TEXT,
			isFavourite INTEGER,
			favouritedAt TEXT
        );
    `)
	_, _ = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS YtHistory(
			sid INTEGER,
            url TEXT PRIMARY KEY,
			thumbnail TEXT,
            title TEXT,
            channel TEXT,
            pubDate TEXT,
            seenAt TEXT,
			isFavourite INTEGER,
			favouritedAt TEXT
        );
    `)

	_, _ = dbInstance.Exec(`
		CREATE TABLE IF NOT EXISTS RssCache(
            url TEXT PRIMARY KEY,
            title TEXT,
            source TEXT,
            category TEXT,
            pubDate TEXT
		);
	`)

	_, _ = dbInstance.Exec(`
        CREATE TABLE IF NOT EXISTS YtCache(
            url TEXT PRIMARY KEY,
			thumbnail TEXT,
            title TEXT,
            channel TEXT,
            pubDate TEXT
        );
    `)
}

func CloseDbInstance() {
	if dbInstance != nil {
		dbInstance.Close()
	}
}
