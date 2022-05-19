package api

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

type Client struct {
	HttpClient *http.Client
	Token      Token
}

const (
	ChimuDownloadURL = "https://api.chimu.moe/v1/download/"
	OAuthURL         = "https://osu.ppy.sh/oauth/token"
	BaseURL          = "https://osu.ppy.sh/api/v2/"
	Best             = "best"
	Firsts           = "firsts"
	Favorite         = "favourite"
	Graveyard        = "graveyard"
	Loved            = "loved"
	MostPlayed       = "most_played"
	Ranked           = "ranked"
	Pending          = "pending"
)

func CreateClient(clientID uint, clientSecret string) Client {
	client := Client{
		HttpClient: http.DefaultClient,
	}
	client.AssignTokenAndCredentials(clientID, clientSecret)

	return client
}

func (c *Client) AssignTokenAndCredentials(clientID uint, clientSecret string) {
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

	res, err := c.HttpClient.Post(OAuthURL, "application/json", bytes.NewBuffer(tokenReqBytes))
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

func (c *Client) GetReq(url string, params map[string]interface{}) ([]byte, error) {
	reqURL := BaseURL + url
	paramsString, err := json.Marshal(params)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to marshal request parameters:\n\t%v", err))
	}

	req, err := http.NewRequest("GET", reqURL, bytes.NewBuffer(paramsString))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to make GET request to %s:\n\t%v", reqURL, err))
	}

	req.Header.Add("Authorization", "Bearer "+c.Token.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil || (res.StatusCode < 200 && res.StatusCode > 299) {
		return nil, errors.New(fmt.Sprintf("Unable to make GET request to %s:\n\t%v", reqURL, err))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read response body from GET request:\n\t%v", err))
	}

	return body, nil
}

func (c *Client) GetUserBeatmapsets(userID uint, beatmapType string, params map[string]interface{}) []Beatmapset {
	var beatmapsets []Beatmapset

	url := fmt.Sprintf("users/%d/beatmapsets/%s", userID, beatmapType)
	if body, err := c.GetReq(url, params); err != nil {
		log.Fatalf("Unable to get user beatmaps:\n\t%v", err)
	} else if err := json.Unmarshal(body, &beatmapsets); err != nil {
		log.Fatalf("Unable to unmarshal JSON array of beatmaps into []Beatmapset:\n\t%v", err)
		return []Beatmapset{}
	}

	return beatmapsets
}

func (c *Client) GetUserBeatmapsetsFromScores(userID uint, scoreType string, params map[string]interface{}) []Beatmapset {
	var scores []Score
	var beatmapsets []Beatmapset

	url := fmt.Sprintf("users/%d/scores/%s", userID, scoreType)
	if body, err := c.GetReq(url, params); err != nil {
		log.Fatalf("Unable to get user beatmaps from scores:\n\t%v", err)
	} else if err := json.Unmarshal(body, &scores); err != nil {
		log.Fatalf("Unable to unmarshal JSON array of scores into []Score:\n\t%v", err)
	}

	for _, score := range scores {
		beatmapsets = append(beatmapsets, score.Beatmapset)
	}

	return beatmapsets
}

func (c *Client) GetUserMostPlayedBeatmapsets(userID uint, params map[string]interface{}) []Beatmapset {
	var beatmapsetPlaycounts []BeatmapPlaycount
	var beatmapsets []Beatmapset

	url := fmt.Sprintf("users/%d/beatmapsets/%s", userID, MostPlayed)
	if body, err := c.GetReq(url, params); err != nil {
		log.Fatalf("Unable to get user most played beatmaps:\n\t%v", err)
	} else if err := json.Unmarshal(body, &beatmapsetPlaycounts); err != nil {
		log.Fatalf("Unable to unmarshal JSON array of beatmaps into []Beatmapset:\n\t%v", err)
	}

	for _, playcount := range beatmapsetPlaycounts {
		beatmapsets = append(beatmapsets, playcount.Beatmapset)
	}

	return beatmapsets
}

func (c *Client) GetUser(userID uint) User {
	var user User

	url := fmt.Sprintf("users/%d", userID)
	if body, err := c.GetReq(url, map[string]interface{}{"key": "id"}); err != nil {
		log.Fatalf("Unable to get user:\n\t%v", err)
	} else if err := json.Unmarshal(body, &user); err != nil {
		log.Fatalf("Unable to unmarshal user %d:\n\t%v", userID, err)
	}

	return user
}

func (c *Client) GetBeatmapsetsForUser(userID uint, beatmapTypes map[string]bool, beatmapCounts map[string]uint, gameMode string) map[uint]Beatmapset {
	fmt.Printf("Fetching beatmapset IDs for %d, this may take a while.\n", userID)
	var beatmapsets = make(map[uint]Beatmapset)

	for beatmapType, include := range beatmapTypes {
		if include {
			beatmapsForType := c.GetBeatmapsetsForType(userID, beatmapType, beatmapCounts[beatmapType], gameMode)

			for _, beatmapset := range beatmapsForType {
				beatmapsets[beatmapset.ID] = beatmapset
			}
		}
	}

	return beatmapsets
}

func (c *Client) GetBeatmapsetsForType(userID uint, beatmapType string, mapCount uint, gameMode string) map[uint]Beatmapset {
	var beatmapsets = make(map[uint]Beatmapset)
	forRange := int(math.Ceil(float64(mapCount) / 100))
	offset := 0

	for i := 0; i < forRange; i++ {
		var chunk []Beatmapset
		if beatmapType == MostPlayed {
			chunk = c.GetUserMostPlayedBeatmapsets(userID, map[string]interface{}{
				"limit":  100,
				"offset": offset,
			})
		} else if beatmapType == Best || beatmapType == Firsts {
			chunk = c.GetUserBeatmapsetsFromScores(userID, beatmapType, map[string]interface{}{
				"mode":   gameMode,
				"limit":  100,
				"offset": offset,
			})
		} else {
			chunk = c.GetUserBeatmapsets(userID, beatmapType, map[string]interface{}{
				"limit":  100,
				"offset": offset,
			})
		}

		for _, beatmapset := range chunk {
			beatmapsets[beatmapset.ID] = beatmapset
		}

		offset += 100
	}

	return beatmapsets
}

func (c *Client) DownloadMaps(beatmapsets map[uint]Beatmapset, outputDir string) {
	mapsDownloaded := 0
	for _, beatmapset := range beatmapsets {
		downloadURL := fmt.Sprintf("%s%d", ChimuDownloadURL, beatmapset.ID)
		resp, err := c.HttpClient.Get(downloadURL)
		if err != nil {
			fmt.Printf("%d failed, please download manually.\n", beatmapset.ID)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%d failed, please download manually.\n", beatmapset.ID)
			continue
		}
		r := regexp.MustCompile("[<>:\"/\\\\|?*]+")
		rawFileName := fmt.Sprintf("%d %s - %s.osz", beatmapset.ID, beatmapset.Artist, beatmapset.Title)
		fileName := r.ReplaceAllString(rawFileName, "")
		err = os.WriteFile(outputDir+"/"+fileName, data, 0777)
		if err != nil {
			fmt.Printf("%d failed, please download manually.\n", beatmapset.ID)
			log.Fatalln(err)
		}
		mapsDownloaded++
		fmt.Printf("Downloaded %d (%d/%d)\n", beatmapset.ID, mapsDownloaded, len(beatmapsets))
		time.Sleep(300 * time.Millisecond) // chimu rate limit is 5rps according to the devs, but it seems it's not that reliable.

	}
	fmt.Printf("Download complete: Managed to download %d/%d maps.\n", mapsDownloaded, len(beatmapsets))
}

func (c *Client) GetBeatmapIDsForRecursiveFavorites(user User, userBeatmapsets map[User]map[uint]Beatmapset, currentDepth uint, maxDepth uint) map[User]map[uint]Beatmapset {
	if _, userExists := userBeatmapsets[user]; userExists {
		return userBeatmapsets
	}

	fmt.Printf("Fetching beatmapset IDs for %s.\n", user.Username)

	beatmapsets := c.GetBeatmapsetsForType(user.ID, Favorite, user.FavoriteBeatmapsetCount, "osu")
	userBeatmapsets[user] = beatmapsets

	depth := currentDepth + 1

	if depth >= maxDepth {
		return userBeatmapsets
	}

	for _, beatmapset := range beatmapsets {
		creator := c.GetUser(beatmapset.UserID)
		userBeatmapsets = c.GetBeatmapIDsForRecursiveFavorites(creator, userBeatmapsets, depth, maxDepth)
	}

	return userBeatmapsets
}
