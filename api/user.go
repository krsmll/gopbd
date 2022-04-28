package api

type User struct {
	ID                       uint   `json:"id"`
	Username                 string `json:"username"`
	FavoriteBeatmapsetCount  uint   `json:"favourite_beatmapset_count"`
	RankedBeatmapsetCount    uint   `json:"ranked_beatmapset_count"`
	LovedBeatmapsetCount     uint   `json:"loved_beatmapset_count"`
	PendingBeatmapsetCount   uint   `json:"pending_beatmapset_count"`
	GraveyardBeatmapsetCount uint   `json:"graveyard_beatmapset_count"`
	BeatmapPlaycountsCount   uint   `json:"beatmap_playcounts_count"`
	ScoresBestCount          uint   `json:"scores_best_count"`
	ScoresFirstCount         uint   `json:"scores_first_count"`
}
