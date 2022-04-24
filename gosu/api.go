package gosu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

const tokenURL = "https://osu.ppy.sh/oauth/token"
const authCodeURL = "https://osu.ppy.sh/oauth/authorize"
const baseURL = "https://osu.ppy.sh/api/v2/"
const tokenName = "oauth_token.json"

type GosuClient struct {
	cliHttp      *http.Client
	Token        string
	clientSecret string
	clientID     string
}

type oauthTokenResponse struct {
	AccessToken    string    `json:"access_token"`
	ExpiresIn      float64   `json:"expires_in"`
	LastAccessTime time.Time `json:"last_access_time"`
	TokenType      string    `json:"token_type"`
}

func RevalidateOAuthToken(clientSecret string, clientID string) (*oauthTokenResponse, error) {
	var ret *oauthTokenResponse
	var err error

	// if OAuth token doesn't exist in file dir
	if _, err = os.Stat(tokenName); os.IsNotExist(err) {
		fmt.Println("OAuth token not detected. requesting new token...")

		// Request a new token
		if ret, err = requestOAuthToken(clientSecret, clientID); err != nil {
			fmt.Println("Error requesting OAuth token:", err)
		}
		ret.LastAccessTime = time.Now()
		writeOAuthToken(tokenName, ret)

		return ret, nil
	} else {
		var previousToken *oauthTokenResponse
		// In this case, the oauth token file already exists
		if previousToken, err = readOAuthToken(tokenName); err != nil {
			fmt.Println("Error in reading OAuth token:", err)
			return nil, err
		}
		// If the token is expired, get a new one (60 second buffer)
		if float64(time.Since(previousToken.LastAccessTime).Seconds())+60 > previousToken.ExpiresIn {
			fmt.Println("OAuth token expired. Requesting new token...")
			if ret, err = requestOAuthToken(clientSecret, clientID); err != nil {
				fmt.Println("Error requesting Oauth token:", err)
			}

			ret.LastAccessTime = time.Now()
			writeOAuthToken(tokenName, ret)
			return ret, nil
		} else {
			return previousToken, nil
		}
	}
}

// contacts the osu api to request a new OAuth token
func requestOAuthToken(clientSecret string, clientID string) (*oauthTokenResponse, error) {
	var ret oauthTokenResponse

	// Make token request
	resp, err := http.PostForm(tokenURL, url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
		"scope":         {"public"},
	})
	if err != nil {
		fmt.Println("Error in getOauthToken POST:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error in reading POST response:", err)
		return nil, err
	}

	// Unmarshal the json into the oauthTokenResponse struct
	if err := json.Unmarshal(body, &ret); err != nil {
		fmt.Println("Error in unmarshaling body:", err)
		return nil, err
	}

	// Timestamp the token
	ret.LastAccessTime = time.Now()
	return &ret, nil
}

func readOAuthToken(path string) (*oauthTokenResponse, error) {
	var ret *oauthTokenResponse

	// If the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s not found", path)
	} else {
		// Read the file
		if body, err := os.ReadFile(path); err != nil {
			fmt.Println("Error reading file:", err)
			return nil, err
		} else {
			// Decode body into ret
			if err := json.Unmarshal(body, &ret); err != nil {
				fmt.Println("Error in unmarshaling body:", err)
				return nil, err
			}
			return ret, nil
		}
	}
}

func writeOAuthToken(path string, token *oauthTokenResponse) error {
	if data, err := json.Marshal(token); err != nil {
		fmt.Println("Error converting struct oauthTokenResponse to json:", err)
		return err
	} else if f, err := os.Create(path); err != nil {
		fmt.Println("Error creating file:", err)
		return err
	} else {
		defer f.Close()
		f.WriteString(string(data))
	}
	return nil
}

func CreateGosuClient(clientSecret string, clientID string) (*GosuClient, error) {
	var oauthToken *oauthTokenResponse
	var err error

	if oauthToken, err = RevalidateOAuthToken(clientSecret, clientID); err != nil {
		fmt.Println("Problem Revalidating Oauth token:", err)
		return nil, err
	}

	var http_client *http.Client = &http.Client{}

	return &GosuClient{cliHttp: http_client, clientSecret: clientSecret, clientID: clientID, Token: oauthToken.AccessToken}, nil
}

// Validates Token and updates client with a valid token
func (c *GosuClient) validateToken() error {
	var oauthToken *oauthTokenResponse
	var err error

	if oauthToken, err = RevalidateOAuthToken(c.clientSecret, c.clientID); err != nil {
		fmt.Println("Problem Revalidating Oauth token:", err)
		return err
	}

	c.Token = oauthToken.AccessToken

	return nil
}

/*
 Send a request to the Gosu client, given:
	`method`: "GET/POST"
	`url`: the HTTP request url, e.g. "beatmaps/lookup"
	`params`: the parameters, in map[string]interface{}{...} format
*/
func (c *GosuClient) DoRequest(method string, url string, params map[string]interface{}) ([]byte, error) {
	// Validate the token
	if err := c.validateToken(); err != nil {
		fmt.Println("Error in validating token:", err)
		return nil, err
	}

	// Create a new request with the osu api url and method
	if args, err := json.Marshal(params); err != nil {
		fmt.Println("Error parsing parameters:", err)
		return nil, err
	} else if req, err := http.NewRequest(method, baseURL+url, bytes.NewBuffer(args)); err != nil {
		fmt.Println("Error creating http request:", err)
		return []byte(""), err
	} else {
		// Add header token
		req.Header.Add("Authorization", "Bearer "+c.Token)
		req.Header.Add("Content-type", "application/json")

		// Send the request
		var resp *http.Response
		if resp, err = c.cliHttp.Do(req); err != nil {
			fmt.Println("Error doing Request:", err)
			return []byte(""), err
		}

		defer resp.Body.Close()

		// Turn the request into a string, and return
		fmt.Println(resp.Body)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error in reading response:", err)
			return []byte(""), err
		}

		return body, nil
	}
}
