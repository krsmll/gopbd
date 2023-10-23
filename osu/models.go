package osu

import "net/http"

type Beatmapset struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type Score struct {
	Beatmapset Beatmapset `json:"beatmapset"`
}

type Token struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type TokenRequest struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}

type Client struct {
	HttpClient *http.Client
	Token      Token
}

type (
	URL         string
	BeatmapType string
	Gamemode    string
)

func GamemodeFromString(s string) Gamemode {
	switch s {
	case "osu":
		return GamemodeOsu
	case "taiko":
		return GamemodeTaiko
	case "fruits":
		return GamemodeCatch
	case "mania":
		return GamemodeMania
	default:
		return GamemodeOsu
	}
}

type User struct {
	ID                       int    `json:"id"`
	Username                 string `json:"username"`
	FavoriteBeatmapsetCount  int    `json:"favourite_beatmapset_count"`
	RankedBeatmapsetCount    int    `json:"ranked_beatmapset_count"`
	LovedBeatmapsetCount     int    `json:"loved_beatmapset_count"`
	PendingBeatmapsetCount   int    `json:"pending_beatmapset_count"`
	GraveyardBeatmapsetCount int    `json:"graveyard_beatmapset_count"`
	BeatmapPlaycountsCount   int    `json:"beatmap_playcounts_count"`
	ScoresBestCount          int    `json:"scores_best_count"`
	ScoresFirstCount         int    `json:"scores_first_count"`
}

func (u *User) GetBeatmapCounts() map[BeatmapType]int {
	return map[BeatmapType]int{
		BeatmapTypeFavorite:   u.FavoriteBeatmapsetCount,
		BeatmapTypeRanked:     u.RankedBeatmapsetCount,
		BeatmapTypeLoved:      u.LovedBeatmapsetCount,
		BeatmapTypePending:    u.PendingBeatmapsetCount,
		BeatmapTypeGraveyard:  u.GraveyardBeatmapsetCount,
		BeatmapTypeMostPlayed: u.BeatmapPlaycountsCount,
		BeatmapTypeBest:       u.ScoresBestCount,
		BeatmapTypeFirsts:     u.ScoresFirstCount,
	}
}
