package handlers

import (
	"net/http"

	"github.com/Ssawa/LinkLetter/web/auth/authentication"
	"github.com/gorilla/mux"
)

// IndexHandlerManager is responsible for, surprise surprise, handling the index of our webpage.
// In my mind however, this includes the routes not only of "/"" but the routes that "/" might
// redirect to, such as "/login". See web/handlers/common.go for a deeper explanation of why
// I'm not totally happy with this BaseHandlerManager thing.
type IndexHandlerManager struct {
	BaseHandlerManager
}

func (manager IndexHandlerManager) index(w http.ResponseWriter, r *http.Request) {
	manager.templator.RenderTemplate(w, "index.tmpl", nil)
}

func (manager IndexHandlerManager) login(w http.ResponseWriter, r *http.Request) {
	manager.templator.RenderTemplate(w, "login.tmpl", nil)
}

func (manager *IndexHandlerManager) InitRoutes(router *mux.Router) http.Handler {
	router.HandleFunc("/", auth.ProtectedFunc(manager.cookies, manager.index))
	router.HandleFunc("/login", manager.login)
	return router
}
