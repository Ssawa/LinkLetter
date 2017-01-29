package web

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestCreateServer(t *testing.T) {
	// We want to move our directory up so that we import
	// our actual templates on Server creation
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("../")

	db, _, _ := sqlmock.New()

	server := CreateServer(
		config.Config{
			SecretKey: "test",
		},
		db,
	)

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	server.Router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

type DummyHandlerManager struct {
	InitializeResourcesWasCalled bool
	InitRoutesWasCalled          bool
	RootHandlerFuncWasCalled     bool
	NestedHandlerFuncWasCalled   bool
	t                            *testing.T
}

func (manager *DummyHandlerManager) InitializeResources(db *sql.DB, cookies *sessions.CookieStore, templator *template.Templator) {
	manager.InitializeResourcesWasCalled = true
	assert.NotNil(manager.t, db)
	assert.NotNil(manager.t, cookies)
	assert.NotNil(manager.t, templator)
}

func (manager *DummyHandlerManager) InitRoutes(router *mux.Router) {
	manager.InitRoutesWasCalled = true

	router.Path("/nested").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manager.NestedHandlerFuncWasCalled = true
	})

	router.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manager.RootHandlerFuncWasCalled = true
	})
}

func TestInitializeManager(t *testing.T) {
	db, _, _ := sqlmock.New()

	router := mux.NewRouter()

	server := Server{
		Router:    router,
		db:        db,
		templator: new(template.Templator),
		cookies:   sessions.NewCookieStore([]byte("secret")),
	}

	testHandlerManager := DummyHandlerManager{false, false, false, false, t}
	server.InitializeManager("/test", &testHandlerManager)
	assert.True(t, testHandlerManager.InitializeResourcesWasCalled)
	assert.True(t, testHandlerManager.InitRoutesWasCalled)

	req := httptest.NewRequest("GET", "/test/", nil)
	resp := httptest.NewRecorder()
	server.Router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	assert.True(t, testHandlerManager.RootHandlerFuncWasCalled)

	req = httptest.NewRequest("GET", "/test/nested", nil)
	resp = httptest.NewRecorder()
	server.Router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	assert.True(t, testHandlerManager.NestedHandlerFuncWasCalled)
}
