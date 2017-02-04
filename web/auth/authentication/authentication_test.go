package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestProtectedFunc(t *testing.T) {
	cookies := sessions.NewCookieStore([]byte("testing"))

	route := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
	wrapped := ProtectedFunc(cookies, route)

	w := httptest.NewRecorder()
	wrapped(w, httptest.NewRequest("GET", "http://localhost/", nil))
	assert.Equal(t, 302, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	w = httptest.NewRecorder()
	wrapped(w, authenticatedRequest(cookies, true))
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	wrapped(w, authenticatedRequest(cookies, false))
	assert.Equal(t, 302, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}

func TestProtectedHandler(t *testing.T) {
	cookies := sessions.NewCookieStore([]byte("testing"))
	mux := http.NewServeMux()
	handler := ProtectedHandler(cookies, mux)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://localhost/", nil))
	assert.Equal(t, 302, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, authenticatedRequest(cookies, true))
	assert.Equal(t, 404, w.Code)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, authenticatedRequest(cookies, false))
	assert.Equal(t, 302, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}
