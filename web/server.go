package web

import (
	"html/template"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server is a general container for the web container
type Server struct {
	Router    *mux.Router
	templates *template.Template
	cookies   *sessions.CookieStore
}

// CreateServer creates a webserver using the supplied config
func CreateServer(conf config.Config) Server {
	server := Server{
		Router:    mux.NewRouter(),
		templates: template.Must(template.ParseFiles(listTemplates()...)),
		cookies:   sessions.NewCookieStore([]byte(conf.SecretKey)),
	}

	server.Router.HandleFunc("/", server.index).Methods("GET")
	server.Router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	return server
}
