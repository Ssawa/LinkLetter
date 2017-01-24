package web

// The structure of the web layer, specifically the way routing is defined, is something
// I agonized over for awhile. You can find tons of tutorials on how to display a simple
// web page but there's surprisingly little out there about best practices for structuring
// a scalable solution for a full featured project, and the few open source examples I
// skimmed through didn't seem too impressive. What we finally ended up with might not
// be the best example itself, but I'm relatively happy with it and thank that it can
// get the job done.
//
// The problems our architecture needs to overcome is:
// 1. Resources such as the database connection and cookie sessions need to be passed
//    to the http handlers for use in generating the response.
// 2. The package structure needs to be able to support a potentially large number of handlers
//    and routes, preferably organized into similar functionality.
// 3. We want to define our routes as clearly and cleanly as possible without repeating
//    ourselves too much or adding too much tedium.
//
// The initial approach was to have all of the handlers be methods of the Server struct.
// This way defining a route would be as simple as:
//
//     server.Router.HandleFunc("/", server.IndexHandler)
//
// This would easily satisfy problems 1 and 3 but completely fails 2. Go dictates that method
// definitions be in the same folder as the struct definition. That would mean every http
// handler (of which there may be dozens) would need to be defined in one folder, the files
// interlaced with the rest of the web source files. We'd be handtied on being able to organize
// our code (putting all login related code into a login folder, for example) and that's just
// not acceptable.
//
// The next approach that was experimented with was using closures to pass resources into handlers.
// Something that looked like:
//
//     func IndexHandler(db *sql.DB, cookies *sessions.CookieStore, templator *template.Templator) func(w http.ResponseWriter, r *http.Request) {
//      	return func(w http.ResponseWriter, r *http.Request) {
//				db.Query("SELECT * FROM TABLE")
//				templator.RenderTemplate(w, "index.tmpl", nil)
//			}
// 		}
//
// We could now define our handlers whereever we want, resolving problems 2, but this is hardly
// clear or clean. I can barely follow it myself. Moreover, every handler now would need to include
// that ugly definition. And what do we do if we need to add another resource to Server down the road?
// We can't pass Server in directly (see InitializeManager documentation for why not) so we would need
// to go into every Handler and change each function signature. That's the definition of tedium. A smaller
// point, but one worth considering, is that in the previous example IndexHandler is *not* actually a
// Handler at all. It simply returns a Handler, which is itself an anonymous function. Maybe it's imagining
// what the stack traces would look like, but something about having our handlers being anonymous functions
// doesn't sit right with me.
//
// In the end, we go with a bit of an amaglamation of the two approaches. We define a common interface
// (see web/handlers/common.go for implementation details) that gets passed the resources and can have
// any number of Handler methods associated with it, leading to code like:
//
//      server.InitializeManager("/", &handlers.IndexHandlerManager{})
//
// This gives us the added bonus of not having to define *every* route explicitly in server.go. For example,
// if we wanted to create a CRUD for "Post" objects, we don't need to define:
//
//      server.Router.HandleFunc("/posts/", ListPosts).Methods("GET")
//      server.Router.HandleFunc("/posts/", CreatePost).Methods("POST")
//      server.Router.HandleFunc("/posts/{id}", UpdatePost).Methods("PUT")
//      ...
//
// We can instead simply define:
//
//      server.InitializeManager("/posts/", &handlers.PostHandlerManager{})
//
// And leave the responsibility of "Post" related routes to PostHandlerManager.
//
// Again, it's not a flawless system, but I hopefully it can pull off the job.

import (
	"database/sql"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/web/handlers"
	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server is a general container for the web container. It holds information
// to handle routing and the resources for the handlers.
type Server struct {
	Router    *mux.Router
	db        *sql.DB
	templator *template.Templator
	cookies   *sessions.CookieStore
}

// CreateServer creates an instance of Server using the supplied config and database connection.
func CreateServer(conf config.Config, db *sql.DB) Server {
	server := Server{
		Router:    mux.NewRouter(),
		db:        db,
		templator: template.CreateDefaultTemplator(),
		cookies:   sessions.NewCookieStore([]byte(conf.SecretKey)),
	}

	// If you're curious why we pass in the pointer reference of the HandlerManager
	// I reccommend the following stack overflow discussion:
	// http://stackoverflow.com/questions/33936081/golang-method-with-pointer-receiver
	// Essentially, because BaseHandlerManager needs its pointer pass in for InitializeResources
	// we need to get the reference here. It's not very intuitive, and I kind of wish Go would
	// make up it's mind about whether I need think about pointers or not.
	server.InitializeManager("/", &handlers.IndexHandlerManager{})

	// This blindly exposes all files in the static folder, so be very careful about what
	// you put in there.
	server.Router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	return server
}

// InitializeManager initializes the handlers.HandlerManager with the server's resource information
// and then registers that handlers.HandlerManager to handle the signified path.
func (server *Server) InitializeManager(path string, manager handlers.HandlerManager) {

	// When I started architecting things I was hoping that a function like this wouldn't be necessary.
	// In an ideal world you'd just create the HandlerManager struct with a reference to Server, and then
	// it would have access to the database and cookies and templator. Unfortunately, this would have meant
	// we would need to import Server into the handlers package and that give us circular dependencies.
	// In C that could have been mitigated with a forward declaration but Go doesn't have them, so instead
	// we need to pass our resources in one at a time.

	manager.InitializeResources(server.db, server.cookies, server.templator)
	server.Router.Handle(path, manager.GetRoutes())
}
