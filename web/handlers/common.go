package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// A HandlerManager is responsible for containing and initializing handlers
// to their specific locations and making web resources available to them
type HandlerManager interface {
	InitializeResources(*sql.DB, *sessions.CookieStore, *template.Templator)
	GetRoutes() http.Handler
}

type BaseHandlerManager struct {
	db        *sql.DB
	cookies   *sessions.CookieStore
	templator *template.Templator
}

func (manager *BaseHandlerManager) InitializeResources(db *sql.DB, cookies *sessions.CookieStore, templator *template.Templator) {
	manager.db = db
	manager.cookies = cookies
	manager.templator = templator
}

func (manager *BaseHandlerManager) GetRoutes() http.Handler {
	return mux.NewRouter()
}
