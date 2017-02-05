package web

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/web/auth/authentication"
	"github.com/Ssawa/LinkLetter/web/auth/oauth2"
	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
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

	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code)

	req = httptest.NewRequest("GET", "/login", nil)
	resp = httptest.NewRecorder()

	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

// #######################################

type dummyHandlerManager struct {
	InitializeResourcesWasCalled bool
	InitRoutesWasCalled          bool
	RootHandlerFuncWasCalled     bool
	NestedHandlerFuncWasCalled   bool
	t                            *testing.T
}

func (manager *dummyHandlerManager) InitializeResources(db *sql.DB, templator *template.Templator, conf *config.Config, login authentication.Login) {
	manager.InitializeResourcesWasCalled = true
	assert.NotNil(manager.t, db)
	assert.NotNil(manager.t, login)
	assert.NotNil(manager.t, templator)
	assert.NotNil(manager.t, conf)
}

func (manager *dummyHandlerManager) InitRoutes(router *mux.Router) http.Handler {
	manager.InitRoutesWasCalled = true

	router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		manager.RootHandlerFuncWasCalled = true
	})

	router.HandleFunc("/nested", func(w http.ResponseWriter, r *http.Request) {
		manager.NestedHandlerFuncWasCalled = true
	})

	return router
}

func TestInitializeManager(t *testing.T) {
	db, _, _ := sqlmock.New()

	router := mux.NewRouter()

	server := Server{
		router:    router,
		db:        db,
		templator: new(template.Templator),
		login:     oauth2.OAuth2Login{},
		conf:      &config.Config{},
	}

	testHandlerManager := dummyHandlerManager{false, false, false, false, t}
	server.initializeManager("/test", &testHandlerManager)
	assert.True(t, testHandlerManager.InitializeResourcesWasCalled)
	assert.True(t, testHandlerManager.InitRoutesWasCalled)

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	assert.True(t, testHandlerManager.RootHandlerFuncWasCalled)

	req = httptest.NewRequest("GET", "/test/nested", nil)
	resp = httptest.NewRecorder()
	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	assert.True(t, testHandlerManager.NestedHandlerFuncWasCalled)
}

// ################################

func middlewareTest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		w.WriteHeader(666)
	})
}

type dummyHandlerManager2 struct {
	RootHandlerFuncWasCalled   bool
	NestedHandlerFuncWasCalled bool
}

func (manager *dummyHandlerManager2) InitializeResources(db *sql.DB, templator *template.Templator, conf *config.Config, login authentication.Login) {
}

func (manager *dummyHandlerManager2) InitRoutes(router *mux.Router) http.Handler {
	router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		manager.RootHandlerFuncWasCalled = true
	})

	router.HandleFunc("/nested", func(w http.ResponseWriter, r *http.Request) {
		manager.NestedHandlerFuncWasCalled = true
	})

	return middlewareTest(router)
}

func TestInitializeManagerAndMiddleware(t *testing.T) {
	router := mux.NewRouter()

	server := Server{
		router: router,
	}

	testHandlerManager := dummyHandlerManager2{}
	server.initializeManager("/test", &testHandlerManager)

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 666, resp.Code)
	assert.True(t, testHandlerManager.RootHandlerFuncWasCalled)

	req = httptest.NewRequest("GET", "/test/nested", nil)
	resp = httptest.NewRecorder()
	server.router.ServeHTTP(resp, req)
	assert.Equal(t, 666, resp.Code)
	assert.True(t, testHandlerManager.NestedHandlerFuncWasCalled)
}
