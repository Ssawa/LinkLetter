package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type IndexHandlerManager struct {
	BaseHandlerManager
}

func (manager IndexHandlerManager) IndexHandler(w http.ResponseWriter, r *http.Request) {
	manager.templator.RenderTemplate(w, "login.tmpl", nil)
}

func (manager *IndexHandlerManager) GetRoutes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", manager.IndexHandler).Methods("GET")
	return router
}
