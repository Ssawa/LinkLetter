package oauth2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	baseAuthURL   = "https://accounts.google.com/o/oauth2/v2/auth"
	baseAccessURL = "https://www.googleapis.com/oauth2/v4/token"
	profileURL    = "https://www.googleapis.com/userinfo/v2/me"
)

type googleProfileData struct {
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	HostedDomain  string `json:"hd"`
	VerifiedEmail bool   `json:"verified_email"`
}

type Google struct {
}

func (google Google) GenerateAuthorizationURL(redirectURI, clientID, scope string) *url.URL {
	u, _ := url.Parse(baseAuthURL)
	query := u.Query()
	query.Set("scope", scope)
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("prompt", "select_account")
	query.Set("response_type", "code")
	u.RawQuery = query.Encode()
	return u
}

func (google Google) ExtractAuthorizationCode(req *http.Request) (string, error) {
	err := req.URL.Query().Get("error")
	if err != "" {
		return "", errors.New(err)
	}

	code := req.URL.Query().Get("code")

	if code == "" {
		return "", fmt.Errorf("Could not extract code from url: %s", req.URL)
	}

	return code, nil
}

func (google Google) GenerateAccessTokenRequest(authorizationCode, redirectURI, clientID, clientSecret string) *http.Request {
	data := url.Values{}
	data.Set("code", authorizationCode)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", baseAccessURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func (google Google) ExtractAccessToken(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Received an invalid response: %d - %s", resp.StatusCode, string(body))
	}

	respBody := struct {
		AccessToken string `json:"access_token"`
	}{}
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	return respBody.AccessToken, nil
}

func (google Google) Authenticate(accessToken string, pattern *regexp.Regexp) (bool, error) {
	profile, err := getProfileData(accessToken)
	if err != nil {
		return false, err
	}

	return pattern.Match([]byte(profile.HostedDomain)), nil
}

func getProfileData(accessToken string) (*googleProfileData, error) {
	req, _ := http.NewRequest("GET", profileURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := googleProfileData{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
