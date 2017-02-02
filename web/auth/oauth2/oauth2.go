package oauth2

import (
	"net/http"
	"net/url"
	"regexp"
)

type OAuth2 interface {
	GenerateAuthorizationURL(redirectURL string) string
	ExtractAuthorizationCode(req *http.Request) (string, error)
	GenerateAccessTokenURL(authorizationCode, redirectURI string) *url.URL
	ExtractAccessToken(resp *http.Response) (string, error)
	Authenticate(accessToken string, pattern *regexp.Regexp) (bool, error)
}
