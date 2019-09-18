package handlers

// This file, along with web/server.go, lays a lot of the ground work for the
// web portion of this application may be structured. The documentation
// in web/server.go is suggested for a more thorough explanation.

import (
	"database/sql"
	"net/http"

	"github.com/cj-dimaggio/LinkLetter/config"
	"github.com/cj-dimaggio/LinkLetter/web/auth/oauth2"
	"github.com/cj-dimaggio/LinkLetter/web/template"
	"github.com/gorilla/mux"
)

// A HandlerManager is responsible for containing and initializing one or more handlers
// and associating them to relevant routes, as well as providing reference to several
// application wide variables, such as the templator, cookies, and database connection.
type HandlerManager interface {

	// InitializeResources is where the web server is expected to pass its resources
	// into the handler, where they can be stored by the implementing struct for
	// use by it's handlers.
	InitializeResources(*sql.DB, *template.Templator, *config.Config, oauth2.OAuth2Login)

	// InitRoutes initializes all routes onto a Handler (most probably a mux.Router)
	// where it can be associated with the main server's router at a path prefix
	// of it's choosing.
	InitRoutes(*mux.Router) http.Handler
}

// BaseHandlerManager implements the basic functionality of a HandlerManager, such as
// the InitializeResources method that sets a reference to application wide resources.
//
// For better or worse, I cut my teeth on OOP. I know it's gotten a bad rap lately
// but it's what I know, it's my natural point of reference. As such, when a language
// like Go, which has very loose OOP constructs, comes around I catch myself trying
// to force my old habits into the new ecosystem. I fear that's what's happening with
// BaseHandlerManager.
//
// From my perspective, this was the situation: I wanted to define a prototype of
// expected, common, behaviors for varied, but functionally related, structs so they
// can be used interchangeably in web.Server function calls. Simple enough, that's
// why Go has interfaces and thus we have HandlerManager. However, I also wanted some
// of these defined behaviors to have a common definition; for instance, InitializeResources
// will probably be the same for most implementations of HandlerManager. In something like C++
// that's simple, define a base class that implements that functionality and subclass it
// for more custom behavior. But this isn't C++, this is Go. Go doesn't support polymorphism,
// a struct simply is what it is and does what it does, which is a very clean, interesting,
// concept but is inconvenient for a poltroon who desperately wants to cling to that which
// is familiar and never challenge his or herself to think in new frames of mind, ie: me.
//
// Go, however, does include the concept of "Anonymous fields", which is actually a very
// clever way of handling object composition. Think, for example, of a 3D vector which looks like:
//
//     type Vector3 struct {
//         x int
//         y int
//         z int
//     }
//
// Now imagine an object that we'd like to have a 3D position, using anonymous fields we can define
// this like so:
//
//     type Object struct {
//         Vector3
//     }
//
// We can then access our vector attributes using either object.Vector3.x, or, for convenience, simply:
// object.x
//
// And herein lays how the cantankerous object-oriented programmer can refuse to adapt to the Go way of
// doing things. Could we not use anonymous fields as a substitute for inheritance as well as composition?
// As a matter of fact you can, it's precisely what we're doing with BaseHandlerManager, but you should
// feel dirty for doing so. For one thing we're completely breaking the OOP philosophy of "Object 1 has an
// Object 2" vs "Object 1 is an Object2". When we say:
//
//     type SomeHandlerManager struct {
//         BaseHandlerManager
//     }
//
// We're, technically, saying SomeHandlerManager *has* a BaseHandlerManager, when in fact we're trying to
// create a relationship where SomeHandlerManager *is an* extension of BaseHandlerManager. But here you can
// make the argument that these "is a/has a" concepts are silly restrictions, and if the code works then it
// works. Instead, the reason I'm so disheartened with this is that we've completely missed an opportunity
// to try to learn how Go, specifically, thinks we can model this functionality. We're essentially denying
// ourselves a chance to expand our programming repertoire by exploring a new paradigm for nothing more than
// the comfortability of doing things the way we've always done it.
//
// Obviously this isn't the end of the world, I mean I'm the one who did it and I can tell you I'm going to
// sleep fine tonight. But it's important to recognize these shortcomings, when they happen to come up,
// so we can't push ourselves a little harder next time. For now, I hope that somebody takes up the task
// of exploring what an alternative approach to this issue might be that's more inline with Go's philosophies,
// as I'd very much be interested in learning it.
type BaseHandlerManager struct {
	db        *sql.DB
	login     oauth2.OAuth2Login
	templator *template.Templator
	conf      *config.Config
}

// InitializeResources handles the base functionality of taking the resource references from the Server and storing
// them for use by our handlers.
func (manager *BaseHandlerManager) InitializeResources(db *sql.DB, templator *template.Templator, conf *config.Config, login oauth2.OAuth2Login) {
	manager.db = db
	manager.templator = templator
	manager.conf = conf
	manager.login = login
}

// InitRoutes doesn't do much here. Actually it does functionally nothing. It only really exists so that BaseHandlerManager
// completely implements HandlerManager and to give a more thorough template to those who may wish to "inherit"
// from BaseHandlerManager.
func (manager *BaseHandlerManager) InitRoutes(router *mux.Router) http.Handler {
	return router
}
