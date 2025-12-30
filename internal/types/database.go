package types

type DatabaseRssItem struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Link         string `json:"link"`
	PubDate      string `json:"pubDate"`
	ReadAt       string `json:"readAt"`
	Source       string `json:"source"`
	Category     string `json:"category"`
	IsFavourite  bool   `json:"isFavourite"`
	FavouritedAt string `json:"favouritedAt"`
}
