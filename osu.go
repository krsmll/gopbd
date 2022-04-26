package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

var oauthURL = "https://osu.ppy.sh/oauth/token"
var apiBaseURL = "https://osu.ppy.sh/api/v2/"

// osu! structs are missing many unneeded fields for the application to operate.
// For the complete list of fields/objects visit: https://osu.ppy.sh/docs/index.html

type Beatmapset struct {
	ID     uint   `json:"id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type BeatmapPlaycount struct {
	Beatmapset Beatmapset `json:"beatmapset"`
}

type User struct {
	ID                       uint   `json:"id"`
	Username                 string `json:"username"`
	FavoriteBeatmapsetCount  uint   `json:"favourite_beatmapset_count"`
	RankedBeatmapsetCount    uint   `json:"ranked_beatmapset_count"`
	LovedBeatmapsetCount     uint   `json:"loved_beatmapset_count"`
	PendingBeatmapsetCount   uint   `json:"pending_beatmapset_count"`
	GraveyardBeatmapsetCount uint   `json:"graveyard_beatmapset_count"`
	BeatmapPlaycountsCount   uint   `json:"beatmap_playcounts_count"`
}

type Token struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   uint   `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type TokenRequest struct {
	ClientID     uint   `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}

type Client struct {
	HttpClient   *http.Client
	Token        Token
	ClientID     uint
	ClientSecret string
}

const (
	FAVOURITE   = "favourite"
	GRAVEYARD   = "graveyard"
	LOVED       = "loved"
	MOST_PLAYED = "most_played"
	RANKED      = "ranked"
	PENDING     = "pending"
)

func CreateClient(clientID uint, clientSecret string) Client {
	client := Client{
		HttpClient: http.DefaultClient,
	}
	err := client.AssignTokenAndCredenitals(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Unable to retrieve Token:\n%s", err)
	}

	return client
}

func (c *Client) AssignTokenAndCredenitals(clientID uint, clientSecret string) error {
	reqToken := TokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
		Scope:        "public",
	}
	tokenReqBytes, err := json.Marshal(reqToken)
	if err != nil {
		log.Fatalf("Unable to marshal TokenRequest:\n%s", err)
	}

	res, err := c.HttpClient.Post(oauthURL, "application/json", bytes.NewBuffer(tokenReqBytes))
	if err != nil {
		log.Fatalf("Unable to retrieve Token from POST request:\n%s", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Unable to read response body from POST request:\n%s", err)
	}

	err = json.Unmarshal(body, &c.Token)
	if err != nil {
		log.Fatalf("Unable to unmarshal Token JSON string to Token struct:\n%s", err)
	}
	c.ClientID = clientID
	c.ClientSecret = clientSecret

	return nil
}

func (c *Client) GetReq(url string, params map[string]interface{}) []byte {
	reqURL := apiBaseURL + url
	paramsString, err := json.Marshal(params)
	if err != nil {
		log.Fatalf("Unable to marshal request parameters:\n%s", err)
	}

	req, err := http.NewRequest("GET", reqURL, bytes.NewBuffer(paramsString))
	if err != nil {
		log.Fatalf("Unable to make GET request to %s:\n%s", reqURL, err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		log.Fatalf("Unable to make GET request to %s:\n%s", reqURL, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Unable to read response body from GET request:\n%s", err)
	}

	return body
}

func (c *Client) GetUserBeatmapsets(userID uint, beatmapType string, params map[string]interface{}) []Beatmapset {
	var beatmapsets []Beatmapset

	url := "users/" + strconv.FormatUint(uint64(userID), 10) + "/beatmapsets/" + beatmapType
	body := c.GetReq(url, params)
	err := json.Unmarshal(body, &beatmapsets)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON array of beatmaps into []Beatmapset:\n%s", err)
		return []Beatmapset{}
	}
	return beatmapsets
}

func (c *Client) GetUserMostPlayedBeatmapsets(userID uint, params map[string]interface{}) []BeatmapPlaycount {
	var beatmapsetPlaycounts []BeatmapPlaycount

	url := "users/" + strconv.FormatUint(uint64(userID), 10) + "/beatmapsets/" + MOST_PLAYED
	body := c.GetReq(url, params)
	err := json.Unmarshal(body, &beatmapsetPlaycounts)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON array of beatmaps into []Beatmapset:\n%s", err)
		return []BeatmapPlaycount{}
	}
	return beatmapsetPlaycounts
}

func (c *Client) GetUser(userID uint) User {
	var user User

	userIDString := strconv.FormatUint(uint64(userID), 10)
	url := "users/" + userIDString
	body := c.GetReq(url, map[string]interface{}{
		"key": "id",
	})
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Fatalf("Unable to unmarshal user %s:\n%s", userIDString, err)
	}

	return user
}

func (c *Client) GetBeatmapsetsForUser(userID uint, beatmapTypes map[string]bool, beatmapCounts map[string]uint) []Beatmapset {
	var beatmapsets []Beatmapset

	for beatmapType, include := range beatmapTypes {
		if include {
			beatmapsets = append(beatmapsets, c.GetBeatmapsetsForType(userID, beatmapType, beatmapCounts[beatmapType])...)
		}
	}

	return beatmapsets
}

func (c *Client) GetBeatmapsetsForType(
	userID uint,
	beatmapType string,
	mapCount uint,
) []Beatmapset {
	var beatmapsets []Beatmapset
	forRange := int(math.Ceil(float64(mapCount) / 100))
	offset := 0

	for i := 0; i < forRange; i++ {
		if beatmapType == MOST_PLAYED {
			playcountChunk := c.GetUserMostPlayedBeatmapsets(userID, map[string]interface{}{
				"limit":  100,
				"offset": offset,
			})
			for _, playcount := range playcountChunk {
				beatmapsets = append(beatmapsets, playcount.Beatmapset)
			}
		} else {
			chunk := c.GetUserBeatmapsets(userID, beatmapType, map[string]interface{}{
				"limit":  100,
				"offset": offset,
			})

			beatmapsets = append(beatmapsets, chunk...)
		}

		offset += 100
	}

	return beatmapsets
}
