package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// LoginHandlerManager is responsible for all routes that log a user in to the site.
type LoginHandlerManager struct {
	BaseHandlerManager
}

func (manager LoginHandlerManager) loginFunc(w http.ResponseWriter, r *http.Request) {
	manager.templator.RenderTemplate(w, "login.tmpl", struct{ OAuth2URL string }{manager.login.GetAuthorizationURL()})
}

func (manager *LoginHandlerManager) InitRoutes(router *mux.Router) http.Handler {
	router.HandleFunc("", manager.loginFunc)
	router.HandleFunc("/auth/oauth2/google", manager.login.AuthorizationCallbackHandler)
	return router
}
