package web

import "github.com/gorilla/mux"
import "net/http"

type Router struct {
	*mux.Router
	Handler http.Handler
}

func (router Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler
	if router.Handler != nil {
		handler = router.Handler
	} else {
		handler = router.Router
	}
	handler.ServeHTTP(w, req)
}
