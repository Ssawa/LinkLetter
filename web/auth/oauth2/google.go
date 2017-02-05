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

	"github.com/Ssawa/LinkLetter/logger"
)

// These are the basic Google APIs we'll be using
const (
	baseAuthURL   = "https://accounts.google.com/o/oauth2/v2/auth"
	baseAccessURL = "https://www.googleapis.com/oauth2/v4/token"
	profileURL    = "https://www.googleapis.com/userinfo/v2/me"
)

// googleProfileData simply holds the google specific information from
// querying their APIs
type googleProfileData struct {
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	HostedDomain  string `json:"hd"`
	VerifiedEmail bool   `json:"verified_email"`
}

// Google is an object in which to assign our OAuth2 method implementations
type Google struct{}

// GenerateAuthorizationURL generates a URL to google's OAuth2 service so that a user can
// give us permission to access their account
func (google Google) GenerateAuthorizationURL(redirectURI, clientID, scope string) string {
	u, _ := url.Parse(baseAuthURL)
	query := u.Query()
	query.Set("scope", scope)
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("prompt", "select_account")
	query.Set("response_type", "code")
	u.RawQuery = query.Encode()
	return u.String()
}

// ExtractAuthorizationCode pulls out the authorization code from the request if it exists
func (google Google) ExtractAuthorizationCode(req *http.Request) (string, error) {
	err := req.URL.Query().Get("error")
	if err != "" {
		logger.Error.Printf("Google authorization error: %s", err)
		return "", errors.New(err)
	}

	code := req.URL.Query().Get("code")

	if code == "" {
		return "", fmt.Errorf("Could not extract code from url: %s", req.URL)
	}

	return code, nil
}

// GenerateAccessTokenRequest generates a request to Google's API to translate our authorization code into a functioning
// access token.
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

// ExtractAccessToken extracts the access token from Google's JSON response.
//
// Now here is one area where I'm not happy with having to generalize our OAuth2 procedure.
// The problem is that google is actually sending us two very useful pieces of data in their
// response that we're unable to make real good use of.
//
// The first is google's "Refresh Token". By default our access token will expire after about
// an hour (this is OAuth2's slightly busted way of trying to deal with replay attacks without
// having to force timestamps and nonces). They however provide a refresh token so that you can
// request a new access token to increase your expiration time. However this is very much
// implementation specific (github, for instance, doesn't have any concept of a refresh token)
// so any reference to it stays out of our OAuth2 interface. Luckily, at the time of writing,
// we're only using our access token once as soon as we get it to get the user's profile info
// and log them in, so we don't really need it. But if we did we would need to save it to
// the google's struct and maintain it internally, which might be hard/impossible.
//
// The other valuable piece of data we lose is that google actually passes us a JWT with all of
// the user's information encoded in it, meaning an additional API call to get profile info is
// not even needed. Again, this is very provider specific so places like Github have no concept
// of it and as such we don't take advantage of it in the OAuth2 interface. We could, theoretically,
// cache this data in the struct and then decode it in "Authenticate" but it seems a little too
// delicate for my case. (If however this were a mission critical application I might consider it.
// I'd actually probably restructure all of our architecture if this were an important production
// piece; but as it is a fun little side project, it's focus is learning and teaching, and the
// coding conventions reflect that even if they are less performant)
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
		logger.Error.Printf("Could not decode json: %s", err)
		return "", err
	}

	return respBody.AccessToken, nil
}

// Authenticate matches the passed in regexp with the user's hosted domain to see if they have access
// to log in.
//
// This is another place where the generalization kind of breaks down. I mean how do we plan to reuse
// this function signature for something like github where they don't even validate email addresses?
// We probably won't be able to and we'll need to revise our entire system. But for now, we're just
// working on our first pass.
func (google Google) Authenticate(accessToken string, pattern *regexp.Regexp) (bool, error) {
	profile, err := getProfileData(accessToken)
	if err != nil {
		logger.Error.Printf("Could not get profile data: %s", err)
		return false, err
	}

	return pattern.Match([]byte(profile.HostedDomain)), nil
}

// getProfileData retrieves a person's profile data from Google's API.
//
// To be honest this could
// very easily (and arguably should be) inlined into Authenticate seeing as nobody else uses it.
// But I'm forseeing a situation where we'll want to refactor our authentication process and
// having this extracted out already might make that easier in the future.
//
// As noted in ExtractAccessToken, this entire function could be made redundant if we simply
// held onto the JWT that google passes to us. But as of right now, to more accurately display
// the OAuth2 process, we do not.
func getProfileData(accessToken string) (*googleProfileData, error) {
	req, _ := http.NewRequest("GET", profileURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error.Printf("Error performing API request: %s", err)
		return nil, err
	}

	data := googleProfileData{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		logger.Error.Printf("Error decoding json: %s", err)
		return nil, err
	}

	return &data, nil
}
