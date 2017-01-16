package web

import "net/http"

func (server Server) index(w http.ResponseWriter, r *http.Request) {
	server.renderTemplate(w, "login.tmpl", struct{}{})
}
