package osu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Official osu! API v2 docs: https://osu.ppy.sh/docs/index.html

const (
	URLChimuDownload URL = "https://api.chimu.moe/v1/download/"
	URLAuth          URL = "https://osu.ppy.sh/oauth/token"
	URLBase          URL = "https://osu.ppy.sh/api/v2/"

	BeatmapTypeBest       BeatmapType = "best"
	BeatmapTypeFirsts     BeatmapType = "firsts"
	BeatmapTypeFavorite   BeatmapType = "favourite"
	BeatmapTypeGraveyard  BeatmapType = "graveyard"
	BeatmapTypeLoved      BeatmapType = "loved"
	BeatmapTypeMostPlayed BeatmapType = "most_played"
	BeatmapTypeRanked     BeatmapType = "ranked"
	BeatmapTypePending    BeatmapType = "pending"

	GamemodeOsu   Gamemode = "osu"
	GamemodeTaiko Gamemode = "taiko"
	GamemodeCatch Gamemode = "fruits"
	GamemodeMania Gamemode = "mania"
)

func NewClient(clientID int, clientSecret string) Client {
	client := Client{
		HttpClient: http.DefaultClient,
	}
	client.setCredentials(clientID, clientSecret)

	return client
}

func (c *Client) setCredentials(clientID int, clientSecret string) {
	reqToken := TokenRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
		Scope:        "public",
	}
	tokenReqBytes, err := json.Marshal(reqToken)
	if err != nil {
		log.Fatalf("Unable to marshal TokenRequest:\n\t%v", err)
	}

	req, err := http.NewRequest("POST", string(URLAuth), bytes.NewBuffer(tokenReqBytes))
	if err != nil {
		log.Fatalf("Unable to make POST request to %s:\n\t%v", string(URLAuth), err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		log.Fatalf("Unable to retrieve Token from POST request:\n\t%v", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Unable to read response body from POST request:\n\t%v", err)
	}

	err = json.Unmarshal(body, &c.Token)
	if err != nil {
		log.Fatalf("Unable to unmarshal Token JSON string to Token struct:\n\t%v", err)
	}
}

func (c *Client) getReq(url string, params map[string]interface{}, osuURL bool) ([]byte, error) {
	paramsString, err := json.Marshal(params)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to marshal request parameters:\n\t%v", err))
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(paramsString))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to make GET request to %s:\n\t%v", url, err))
	}

	if osuURL {
		req.Header.Add("Authorization", "Bearer "+c.Token.AccessToken)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
	}

	res, err := c.HttpClient.Do(req)
	if err != nil || (res.StatusCode < 200 && res.StatusCode > 299) {
		return nil, errors.New(fmt.Sprintf("Unable to make GET request to %s:\n\t%v", url, err))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read response body from GET request:\n\t%v", err))
	}

	return body, nil
}

func (c *Client) GetUser(userID int) User {
	var user User

	url := fmt.Sprintf("%susers/%d", URLBase, userID)
	if body, err := c.getReq(url, map[string]interface{}{"key": "id"}, true); err != nil {
		log.Fatalf("Unable to get user:\n\t%v", err)
	} else if err := json.Unmarshal(body, &user); err != nil {
		log.Fatalf("Unable to unmarshal user %d:\n\t%v", userID, err)
	}

	return user
}

func (c *Client) getUserBeatmapsets(userID int, beatmapType BeatmapType, gamemode Gamemode, mapCount int) []Beatmapset {
	var beatmapsets []Beatmapset
	var reqURL string
	forRange := int(math.Ceil(float64(mapCount) / 100))
	offset := 0
	params := map[string]interface{}{
		"limit":  100,
		"offset": offset,
	}

	switch beatmapType {
	case BeatmapTypeBest, BeatmapTypeFirsts:
		reqURL = fmt.Sprintf("%susers/%d/scores/%s", URLBase, userID, beatmapType)
		params["mode"] = gamemode
	default:
		reqURL = fmt.Sprintf("%susers/%d/beatmapsets/%s", URLBase, userID, beatmapType)
	}

	for i := 0; i < forRange; i++ {
		body, err := c.getReq(reqURL, params, true)
		if err != nil {
			log.Fatalf("Unable to get user beatmaps:\n\t%v", err)
		}

		beatmapsets = unmarshalForBeatmapType(body, beatmapType)
	}

	return beatmapsets
}

func (c *Client) StartGatheringBeatmapsets(ch chan int, userID int, beatmapTypes map[BeatmapType]bool, mapCounts map[BeatmapType]int, gamemode Gamemode) {
	fmt.Println("Gathering beatmapsets...")
	for beatmapType, include := range beatmapTypes {
		if !include {
			continue
		}
		beatmapsets := c.getUserBeatmapsets(userID, beatmapType, gamemode, mapCounts[beatmapType])
		for _, beatmapset := range beatmapsets {
			ch <- beatmapset.ID
		}
	}
	fmt.Println("Done gathering beatmapsets. Please wait for downloads to finish.")
}

func (c *Client) Download(ch chan int, outputDir string, beatmapsetCount int) {
	mapsDownloaded := 0
	mapsFailed := 0
	mapsLeft := beatmapsetCount
	for mapsLeft != 0 {
		select {
		case beatmapsetID := <-ch:
			downloadURL := fmt.Sprintf("%s%d", URLChimuDownload, beatmapsetID)
			body, err := c.getReq(downloadURL, nil, false)
			if err != nil {
				fmt.Printf("%d failed, please download manually.\n", beatmapsetID)
				mapsFailed++
				continue
			}

			r := regexp.MustCompile("[<>:\"/\\\\|?*]+")
			rawFileName := fmt.Sprintf("%d.osz", beatmapsetID)
			fileName := r.ReplaceAllString(rawFileName, "")
			err = os.WriteFile(outputDir+"/"+fileName, body, 0777)
			if err != nil {
				fmt.Printf("%d failed, please download manually.\n", beatmapsetID)
				log.Fatalln(err)
			}
			mapsDownloaded++
			mapsLeft--
			fmt.Printf("Downloaded %d (%d/%d)\n", beatmapsetID, mapsDownloaded, beatmapsetCount)
			time.Sleep(350 * time.Millisecond) // chimu rate limit is 5rps according to the devs, 300ms just in case.
		}
	}

	fmt.Printf("Successfully downloaded %d/%d beatmapsets. \n", mapsDownloaded)
}

func unmarshalForBeatmapType(body []byte, beatmapType BeatmapType) []Beatmapset {
	var beatmapsets []Beatmapset

	switch beatmapType {
	case BeatmapTypeBest, BeatmapTypeFirsts, BeatmapTypeMostPlayed:
		var scores []Score
		if err := json.Unmarshal(body, &scores); err != nil {
			log.Fatalf("Unable to unmarshal JSON array of scores into []Score:\n\t%v", err)
		}

		for _, score := range scores {
			beatmapsets = append(beatmapsets, score.Beatmapset)
		}
	default:
		if err := json.Unmarshal(body, &beatmapsets); err != nil {
			log.Fatalf("Unable to unmarshal JSON array of beatmapsets into []Beatmapset:\n\t%v", err)
		}
	}

	return beatmapsets
}
