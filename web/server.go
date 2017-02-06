package web

// The structure of the web layer, specifically the way routing is defined, is something
// I agonized over for awhile. You can find tons of tutorials on how to display a simple
// web page but there's surprisingly little out there about best practices for structuring
// a scalable solution for a full featured project, and the few open source examples I
// skimmed through didn't seem too impressive. What we finally ended up with might not
// be the best example itself, but I'm relatively happy with it and think that it can
// get the job done.
//
// The problems our architecture needs to overcome is:
// 1. Resources such as the database connection and cookie sessions need to be passed
//    to the http handlers for use in generating the response.
// 2. The package structure needs to be able to support a potentially large number of handlers
//    and routes, preferably organized by similar functionality.
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
//      server.Router.HandleFunc("/posts/{id}", GetPost).Methods("GET")
//      server.Router.HandleFunc("/posts/{id}", UpdatePost).Methods("PUT")
//      server.Router.HandleFunc("/posts/{id}", DeletePost).Methods("DELETE")
//      ...
//
// We can instead simply define:
//
//      server.InitializeManager("/posts/", &handlers.PostHandlerManager{})
//
// And leave the responsibility of "Post" related routes entirely to PostHandlerManager.
//
// Again, it's not a flawless system, but hopefully it can pull off the job.
//
// Now, with my thought process out of the way, let's address the 10,000 pound elephant in the room:
// why not just use a goddamn global? You may be thinking I'm being frustratingly obtuse, over-engineering
// a fundamentally simple scenerio just to satisfy an irrational, internal voice that constantly shouts
// down at me, like Yahweh to Moses on that cloudy precipice of Mount Sinai, "Thou shalt be stateless!
// Thou shalt not put database connections into global variables!" You'd only be partially correct.
// Look, I realize that it's not a huge deal, I'm even willing to play along under the right
// circumstances (see logger/logger.go for an example of global variables), but I submit that having standards
// is more important than what, exactly, those standards actually are. While these ubiquitous programming
// "rules" are sometimes questionable, they provide helpful constraints. It forces us to to exercise
// our grey matter, approach problem solving with thoughtfulness and creativity, and, most importantly to me,
// maintain a respect for programming as a craft, in and of itself, rather than a means to an end. Arbitrary
// and hand-wavy, I known, but as David Foster Wallace so succinctly described for us, "...in the day-to-day
// trenches of adult existence, banal platitudes can have life-or-death importance."
//
// Don't, however, fall into the trap of thinking of these standards as rules, that one can never ignore
// in favor of exploring one's own path. Lest we forget what that other literary giant, Neruda, told us:
// "Those who shun the 'bad taste' of things will fall flat on the ice.", I only suggest that if one seeks,
// "...the poetry we search for: worn with the hand's obligations, as by acids, steeped in sweat and in smoke,
// smelling of the lilies and urine, spattered diversely by the trades that we live by, inside the law or
// beyond it," one must first know the laws of their trade.
//
// In other words: no freaking globals, people.

import (
	"database/sql"
	"net/http"

	"fmt"

	"regexp"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/logger"
	"github.com/Ssawa/LinkLetter/web/auth/oauth2"
	"github.com/Ssawa/LinkLetter/web/handlers"
	"github.com/Ssawa/LinkLetter/web/template"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server is a general container for the web container. It holds information
// to handle routing and the resources for the handlers.
type Server struct {
	router    *mux.Router
	db        *sql.DB
	templator *template.Templator
	cookies   *sessions.CookieStore
	conf      *config.Config
	login     oauth2.OAuth2Login
}

// CreateServer creates an instance of Server using the supplied config and database connection.
func CreateServer(conf config.Config, db *sql.DB) Server {
	cookiesStore := sessions.NewCookieStore([]byte(conf.SecretKey))
	pattern, err := regexp.Compile(conf.AuthorizationPattern)
	if err != nil {
		logger.Error.Printf("Unable to compile the string: '%s' into a regular expression. Aborting for security: %s", conf.AuthorizationPattern, err)
		panic(err)
	}

	server := Server{
		router:    mux.NewRouter(),
		db:        db,
		templator: template.CreateDefaultTemplator(),
		cookies:   cookiesStore,
		conf:      &conf,
		login: oauth2.OAuth2Login{
			ClientID:             conf.GoogleClientID,
			ClientSecret:         conf.GoogleClientSecret,
			Scope:                "email",
			RedirectURL:          fmt.Sprintf("%s/login/auth/oauth2/google", conf.URLBase),
			AuthorizationPattern: pattern,
			Cookies:              cookiesStore,
			OAuth2Provider:       oauth2.Google{},
		},
	}

	if !server.login.ShouldAuthenticate() {
		logger.Warning.Printf("You configuration does not support authentication so it will be disabled. This is fine for development purposes " +
			"but is a serious security concern if this is happening on production. If you wish to enable authentication than please update your configuration.")
	}

	server.defineRoutes()

	return server
}

// defineRoutes is used for defining the routes you'd like our Server to serve
func (server *Server) defineRoutes() {
	// This blindly exposes all files in the static folder, so be very careful about what
	// you put in there. I'd also like this line to demonstrate that our system is not
	// dependent on handlers.HandlerManager for routes. HandlerManager is a tool to help us,
	// but a handler is a handler and we can use whatever we want to define our routes.
	server.router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	// Keep in mind, when setting routes, that gorilla/mux will match with the first
	// valid path it finds, not necessarily the most specific. So:
	//     server.router.PathPrefix("/test")
	//     server.router.PathPrefix("/testroute")
	//
	// will match with "/test" for "/testroute/something" before "/testroute". To solve this
	// simply define routes in descending specificity, such as:
	//     server.router.PathPrefix("/testroute")
	//     server.router.PathPrefix("/test")
	//
	// In the future we may want to include a helper function that tries to clean up our route
	// orders for us.

	// If you're curious why we pass in the pointer reference of the HandlerManager
	// I reccommend the following stack overflow discussion:
	// http://stackoverflow.com/questions/33936081/golang-method-with-pointer-receiver
	// Essentially, because BaseHandlerManager needs its pointer passed in for InitializeResources
	// we need to get the reference here. It's not very intuitive, and I kind of wish Go would
	// make up it's mind about whether we need think about pointers or not.
	server.initializeManager("/login", &handlers.LoginHandlerManager{})
	server.initializeManager("/", &handlers.IndexHandlerManager{})
}

// InitializeManager initializes the handlers.HandlerManager with the server's resource information
// and then registers that handlers.HandlerManager to handle the signified path.
func (server *Server) initializeManager(prefix string, manager handlers.HandlerManager) {

	// When I started architecting things I was hoping that a function like this wouldn't be necessary.
	// In an ideal world you'd just create the HandlerManager struct with a reference to Server, and then
	// it would have access to the database and cookies and templator. Unfortunately, this would have meant
	// we would need to import Server into the handlers package and that give us circular dependencies.
	// In C that could have been mitigated with a forward declaration but Go doesn't have them, so instead
	// we need to pass our resources in one at a time.

	manager.InitializeResources(server.db, server.templator, server.conf, server.login)

	// The following few lines of code are the end result of a day of exploring gorilla/mux's subrouter logic
	// and I'm fairly confident that, with the library as it is at the time of writing, this is about as
	// elegant as we can hope to get it. So why so much ado?
	//
	// The problem is that Mux's subrouters don't work well with http middleware. If, for instance, a
	// HandlerManager wanted to wrap all of it's routes with the authentication middleware (as the majority of
	// our routes most likely will need to be), each function would need to be wrapped individually and
	// explicitly with something like:
	//     auth.ProtectedFunc(manager.cookies, manager.handleRoute())
	//
	// which is obviously tedious, ugly, and error prone. The reason is that mux Subrouter's aren't treated
	// as routers at all by gorilla/mux. Their "ServerHTTP" functions are never actually called, as one may
	// intuitively believe. Subrouters are only used for their "Match" method so that their routes may be
	// called directly. And because Subrouters aren't, in fact, routers at all (responsible for, you know,
	// routing) but only a glorified map, no matter how many ways you try to hack it (and hack it I have tried),
	// trying to wrap a piece of middleware around them will just not work.
	//
	// The solution came inspired by the very helpful gist: https://gist.github.com/danesparza/eb3a63ab55a7cd33923e
	// Let's walk through the trick line by line:

	// Here we create a new Subrouter, but notice that it is not a subrouter from our server.router, we create
	// a completely new router object first. PathPrefix and Subrouter have a funny relationship. PathPrefix simply
	// filters. It matches if the prefix condition is met and invokes its handler, it is not responsible for actually
	// stripping that prefix for subsequent routes. Meaning if you have something like:
	//
	//     server.router.PathPrefix("/test").Handler(handler)
	//
	// Path prefix is simply saying to check against the handler if a request like "/test/nested" comes in. That
	// handler, however, still must account for the full route ("/test/nested", not just "/nested") otherwise
	// you'll get a 404.
	//
	// Subrouter provides the other side of the transaction, it simply modifies the handler that PathPrefix forwards
	// to so that you don't need to repeat the prefix "/test" in your route definitions.
	//
	// So when we call:
	//     newRouter := mux.NewRouter().PathPrefix(prefix).Subrouter()
	//
	// newRouter is simply a handler that will strip off that prefix so that it's routes do not need to define
	// it explicitly. (Here we're not really using PathPrefix for its filtering functionality, because of the way
	// mux is written Subrouter just needs it to know what prefix it is to operate on)
	newRouter := mux.NewRouter().PathPrefix(prefix).Subrouter()

	// InitRoutes takes a *mux.Router and returns a simple interface http.Handler (of which *mux.Router implements).
	// This means the manager can take that router (which, again, has already been configured to handle the prefix),
	// set its routes and then simply return the router again, or it can set it's routes, wrap the router in a piece
	// of middleware, and return the wrapped function. The manager could also simply ignore the router and return
	// whatever http.Handler it wants; here we make the assumption that the handler knows what it's doing.
	handler := manager.InitRoutes(newRouter)

	// Now we finally register our new http.Handler with our server's routers. Remember, PathPrefix simply filters
	// and passes on to it's handler. But now, because we've gone through this whole process, that handler now
	// also knows how to handle its prefix
	server.router.PathPrefix(prefix).Handler(handler)
}

// Route prepares the server's routes and return an http.handler on which to serve http requests.
//
// Originally "Server.router" was simply a public member that would be initialized in CreateServer
// and then passed into something like "http.ListenAndServe". However when working on the Authentication
// middleware I found that it was convenient to wrap router in a kind of getter so that I could ensure
// all routes were handled the way we wanted. Furthermore I appreciated the implicit conversion from
// a mux.Router to a simple http.Handler, thus isolating user's of our struct from the non standard lib
// implementation of our functionality. So even though we've since shuffled around the Authentication
// middleware so that it no longer gets initialized here, I decided to keep router private and instead
// restrict access to it through public functions where we have a bit more control.
//
// I was also playing around with the idea of moving defineRoutes() in here, under the belief
// that it was more appropriate under a route focused function but this idea was quickly scrapped because
// it became quickly clear that because Server's only real goal is to, you know, serve http requests, keeping
// defineRoutes() out of the constructure would only serve to cause Server to be constructed in an incomplete
// and unusable state.
func (server *Server) Route() http.Handler {
	return server.router
}
