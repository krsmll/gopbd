package api

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
