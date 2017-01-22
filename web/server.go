package web

import (
	"database/sql"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/web/handlers"
	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server is a general container for the web container
type Server struct {
	Router    *mux.Router
	db        *sql.DB
	templator *template.Templator
	cookies   *sessions.CookieStore
}

// CreateServer creates a webserver using the supplied config
func CreateServer(conf config.Config, db *sql.DB) Server {
	server := Server{
		Router:    mux.NewRouter(),
		db:        db,
		templator: template.CreateDefaultTemplator(),
		cookies:   sessions.NewCookieStore([]byte(conf.SecretKey)),
	}

	// http://stackoverflow.com/questions/33936081/golang-method-with-pointer-receiver
	server.InitializeManager("/", &handlers.IndexHandlerManager{})
	server.Router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	return server
}

func (server *Server) InitializeManager(path string, manager handlers.HandlerManager) {
	manager.InitializeResources(server.db, server.cookies, server.templator)
	server.Router.Handle(path, manager.GetRoutes())
}
