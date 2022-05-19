package api

type Beatmapset struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type BeatmapPlaycount struct {
	Beatmapset Beatmapset `json:"beatmapset"`
}

type Score struct {
	Beatmapset Beatmapset `json:"beatmapset"`
}
