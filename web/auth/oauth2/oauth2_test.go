package oauth2

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"net/http/httptest"

	"github.com/cj-dimaggio/LinkLetter/testhelpers"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestShouldAuthenticate(t *testing.T) {
	login := OAuth2Login{
		ClientID:     "",
		ClientSecret: "",
	}
	assert.False(t, login.ShouldAuthenticate())

	login = OAuth2Login{
		ClientID:     "a",
		ClientSecret: "",
	}
	assert.False(t, login.ShouldAuthenticate())

	login = OAuth2Login{
		ClientID:     "",
		ClientSecret: "b",
	}
	assert.False(t, login.ShouldAuthenticate())

	login = OAuth2Login{
		ClientID:     "a",
		ClientSecret: "b",
	}
	assert.True(t, login.ShouldAuthenticate())

}

// ==================================

type testOAuth2Provider struct {
	ExtractAuthorizationCodeError bool
	ExtractAccessTokenError       bool
	AuthenticateError             bool
	AuthenticationResult          bool

	GenerateAuthorizationURLCalled   bool
	ExtractAuthorizationCodeCalled   bool
	GenerateAccessTokenRequestCalled bool
	ExtractAccessTokenCalled         bool
	AuthenticateCalled               bool
}

func (oauth2 *testOAuth2Provider) GenerateAuthorizationURL(redirectURL, clientID, scope string) string {
	oauth2.GenerateAuthorizationURLCalled = true
	return "http://localhost"
}

func (oauth2 *testOAuth2Provider) ExtractAuthorizationCode(req *http.Request) (string, error) {
	oauth2.ExtractAuthorizationCodeCalled = true
	if oauth2.ExtractAuthorizationCodeError {
		return "", errors.New("ExtractAuthorizationCode Errored")
	}
	return "code", nil
}

func (oauth2 *testOAuth2Provider) GenerateAccessTokenRequest(authorizationCode, redirectURI, clientID, clientSecret string) *http.Request {
	oauth2.GenerateAccessTokenRequestCalled = true
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	return req
}

func (oauth2 *testOAuth2Provider) ExtractAccessToken(resp *http.Response) (string, error) {
	oauth2.ExtractAccessTokenCalled = true
	if oauth2.ExtractAccessTokenError {
		return "", errors.New("ExtractAccessToken Errored")
	}
	return "token", nil
}

func (oauth2 *testOAuth2Provider) Authenticate(accessToken string, pattern *regexp.Regexp) (bool, error) {
	oauth2.AuthenticateCalled = true
	if oauth2.AuthenticateError {
		return false, errors.New("Authenticate Errored")
	}
	return oauth2.AuthenticationResult, nil
}

func authorizationCallbackHandlerRun(extractAuthorizationCodeError, extractAccessTokenError, authenticateError, authenticationResult bool) (*httptest.ResponseRecorder, testOAuth2Provider) {
	transport := testhelpers.FakeTransport(func(req *http.Request) (*http.Response, error) {
		resp := httptest.NewRecorder()
		return resp.Result(), nil
	})
	defer transport.Close()

	provider := testOAuth2Provider{
		ExtractAuthorizationCodeError: extractAuthorizationCodeError,
		ExtractAccessTokenError:       extractAccessTokenError,
		AuthenticateError:             authenticateError,
		AuthenticationResult:          authenticationResult,
	}

	login := OAuth2Login{
		Cookies:        sessions.NewCookieStore([]byte("test")),
		OAuth2Provider: &provider,
	}

	w := httptest.NewRecorder()
	login.AuthorizationCallbackHandler(w, httptest.NewRequest("GET", "http://localhost", nil))
	return w, provider
}

func TestAuthorizationCallbackHandler(t *testing.T) {
	w, provider := authorizationCallbackHandlerRun(false, false, false, true)

	assert.True(t, provider.ExtractAuthorizationCodeCalled)
	assert.True(t, provider.GenerateAccessTokenRequestCalled)
	assert.True(t, provider.ExtractAccessTokenCalled)
	assert.True(t, provider.AuthenticateCalled)
	assert.NotEmpty(t, w.Header().Get("Set-Cookie"))
	assert.Equal(t, 302, w.Code)

	w, provider = authorizationCallbackHandlerRun(false, false, false, false)

	assert.True(t, provider.ExtractAuthorizationCodeCalled)
	assert.True(t, provider.GenerateAccessTokenRequestCalled)
	assert.True(t, provider.ExtractAccessTokenCalled)
	assert.True(t, provider.AuthenticateCalled)
	assert.Empty(t, w.Header().Get("Set-Cookie"))
	assert.Equal(t, 403, w.Code)

	w, provider = authorizationCallbackHandlerRun(true, false, false, true)
	assert.True(t, provider.ExtractAuthorizationCodeCalled)
	assert.False(t, provider.GenerateAccessTokenRequestCalled)
	assert.False(t, provider.ExtractAccessTokenCalled)
	assert.False(t, provider.AuthenticateCalled)
	assert.Empty(t, w.Header().Get("Set-Cookie"))
	assert.Equal(t, 500, w.Code)

	w, provider = authorizationCallbackHandlerRun(false, true, false, true)
	assert.True(t, provider.ExtractAuthorizationCodeCalled)
	assert.True(t, provider.GenerateAccessTokenRequestCalled)
	assert.True(t, provider.ExtractAccessTokenCalled)
	assert.False(t, provider.AuthenticateCalled)
	assert.Empty(t, w.Header().Get("Set-Cookie"))
	assert.Equal(t, 500, w.Code)

	w, provider = authorizationCallbackHandlerRun(false, false, true, true)
	assert.True(t, provider.ExtractAuthorizationCodeCalled)
	assert.True(t, provider.GenerateAccessTokenRequestCalled)
	assert.True(t, provider.ExtractAccessTokenCalled)
	assert.True(t, provider.AuthenticateCalled)
	assert.Empty(t, w.Header().Get("Set-Cookie"))
	assert.Equal(t, 500, w.Code)
}
