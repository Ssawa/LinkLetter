package authentication

import (
	"net/http"
	"testing"

	"net/http/httptest"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

type sessionData map[interface{}]interface{}

func authenticatedRequest(cookies *sessions.CookieStore, authenticated bool) *http.Request {
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	token, _ := securecookie.EncodeMulti("session", sessionData{"isAuthenticated": authenticated}, cookies.Codecs...)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})
	return req
}

func TestIsAuthenticated(t *testing.T) {
	cookies := sessions.NewCookieStore([]byte("testing"))

	auth, err := IsAuthenticated(authenticatedRequest(cookies, true), cookies)
	assert.Nil(t, err)
	assert.True(t, auth)

	auth, err = IsAuthenticated(authenticatedRequest(cookies, false), cookies)
	assert.Nil(t, err)
	assert.False(t, auth)

	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	auth, err = IsAuthenticated(req, cookies)
	assert.Nil(t, err)
	assert.False(t, auth)
}

func TestAuthProtected(t *testing.T) {
	cookies := sessions.NewCookieStore([]byte("testing"))
	mux := http.NewServeMux()
	handler := AuthProtected(cookies, mux)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://localhost/", nil))
	resp := w.Result()
	assert.Equal(t, 302, resp.StatusCode)
	assert.Equal(t, "/login", resp.Header.Get("Location"))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, authenticatedRequest(cookies, true))
	resp = w.Result()
	assert.Equal(t, 404, resp.StatusCode)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, authenticatedRequest(cookies, false))
	resp = w.Result()
	assert.Equal(t, 302, resp.StatusCode)
	assert.Equal(t, "/login", resp.Header.Get("Location"))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://localhost/login", nil))
	resp = w.Result()
	assert.Equal(t, 404, resp.StatusCode)
}
