package database

import "time"

type DatabaseFeedItem struct {
	ID       int
	Title    string
	Link     string
	PubDate  time.Time
	Source   string
	Category string
}

type DatabaseFeedReadItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	PubDate     string `json:"pubDate"`
	ReadAt      string `json:"readAt"`
	Source      string `json:"source"`
	Category    string `json:"category"`
	IsFavourite bool   `json:"isFavourite"`
}

type FavouriteRSSItem struct {
	Title        string `json:"title"`
	Link         string `json:"link"`
	PubDate      string `json:"pubDate"`
	FavouritedAt string `json:"favouritedAt"`
	Source       string `json:"source"`
	Category     string `json:"category"`
}

type DatabaseYTItem struct {
	ID        int
	Title     string
	Link      string
	PubDate   time.Time
	Channel   string
	ThumbNail string
}

type DatabaseYTSeenItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	PubDate     string `json:"pubDate"`
	Channel     string `json:"channel"`
	ThumbNail   string `json:"thumbnail"`
	SeenAt      string `json:"watchedAt"`
	IsFavourite bool   `json:"isFavourite"`
}

type DatabaseYTFavourite struct {
	Title        string `json:"title"`
	Link         string `json:"link"`
	PubDate      string `json:"pubDate"`
	Channel      string `json:"channel"`
	ThumbNail    string `json:"thumbnail"`
	FavouritedAt string `json:"favouritedAt"`
}
