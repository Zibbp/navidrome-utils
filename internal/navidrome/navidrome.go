package navidrome

type Playlist struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Tracks      []Track `json:"tracks"`
}

type Track struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Duration int64  `json:"duration"`
	ISRC     string `json:"isrc"`
}
