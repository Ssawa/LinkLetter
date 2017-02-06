package authentication

import (
	"net/http"

	"github.com/Ssawa/LinkLetter/logger"
	"github.com/gorilla/sessions"
)

const (
	loginPage         string = "/login"
	sessionName       string = "session"
	authenticationKey string = "isAuthenticated"
)

func LogInUser(cookies *sessions.CookieStore, req *http.Request, w http.ResponseWriter) (err error) {
	session, err := cookies.Get(req, sessionName)
	if err == nil {
		session.Values[authenticationKey] = true
		session.Save(req, w)
	}
	return
}

type Login interface {
	ShouldAuthenticate() bool
	GetCookies() *sessions.CookieStore
}

// This package handles the brunt work of our authentication as well as providing a few
// helper methods. The general process for authentication will be:
//
//     * If there is no (or invalid) session cookie in the request redirect to login page
//     * Perform login over OAuth2
//     * Set a session cookie saying user is authenticated
//
// See? Simple. The very important thing to note is that out cookies will all be signed
// cryptographically with a secret key that only we know, thus preventing a user from
// tampering with it. Using a cookie to handle session information has become slightly
// frowned upon I realize, and for arguably good reason. For instance, if you want to
// invalidate a cookie you'd either need to wait for it to expire or change your secret
// key (which would invalidate *everyones* cookies). As such the much more accepted
// solution is to instead store your sessions in the database and instead only use a
// cookie to reference a session key, this way you can easily invalidate individual
// sessions by simply deleteing them. I'll keep beating the "this is a simple, fun side
// project" drum though and say that this added level of maintenance is not strictly
// necessary for us and might only server to obfuscate the process.
//
// (Also, if I'm being brutally honest, the cheapskate in me is also acutely aware that
// Heroku's free tier for postgres limits the number of rows you can have in a database
// and if I can keep this thing running for free for just a little longer than that's just
// gravy.)

// IsAuthenticated determines whether the user's request has a session cookie and if it
// labels them as authenticated.
func IsAuthenticated(req *http.Request, cookies *sessions.CookieStore) (bool, error) {
	session, err := cookies.Get(req, sessionName)
	if err != nil {
		logger.Error.Printf("Encountered error while getting session: %s", err)
		return false, err
	}

	val := session.Values[authenticationKey]
	if authenticated, ok := val.(bool); ok {
		return authenticated, nil
	}

	return false, nil
}

// ProtectedFunc is an http middleware that wraps an http.handlerfunc and checks if a user is authenticated. If the user
// isn't then he/she is redirected to /login, if the user is, then the request continues
// normally
func ProtectedFunc(login Login, wrap func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	if !login.ShouldAuthenticate() {
		return wrap
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, err := IsAuthenticated(r, login.GetCookies())

		if err != nil {
			// I noticed an interesting problem where, because I had been developing another
			// server on the same port I was testing this application on, I just so happened
			// to have a cookie set to our same "sessionName" and it was causing me 500 errors
			// because mux.SecureCookies couldn't decode this unknown format. For this reason,
			// instead of sending out a 500 error as we originally had it, we'll just treat
			// it as if it were an unauthenticated state but not fault. This way, if the user
			// chooses to sign in naturally, their bogus cookie will instead just be overwritten.
			// Obviously the chances of this coming up "in the field" is unlikely, but as it's
			// something that came up already, it would be good not to regress
			logger.Error.Printf("Was unable to determine if user is authenticated, their '%s' cookie may be malformed: %s", sessionName, err)
		}

		if !auth {
			// I kind of think this should be a 303 ("See Other") rather than
			// a 302 ("Found"), but after doing some research it looks like 302s
			// are the standard for cases like this (I believe it's what google uses)
			// so we'll just go with that.
			http.Redirect(w, r, loginPage, 302)
			return
		}

		wrap(w, r)
	})
}

// ProtectedHandler is a piece of middleware that wraps an http.handler with the behavior
// of ProtectedFunc
//
// I struggled a lot with how best to write this middleware, whether it should be handlerfunc
// or handler based. It was originally just handlefunc, which meant that you would need to
// set it explicitly for every route you wanted protected by authentication. However I soon
// realized that, for our particular application, almost every route is going to require authentication
// and forcing it to be explicit would only increase the chance of someone accidentally forgetting
// to add it to a route and it thus being exposed. So to mitigate this I made it a piece of
// http.handler middleware and wrapped the entire application with a few conditionals for things
// like the login page itself and static routes. However this felt (and was) incredibly kludgy
// so I ripped it out and started trying to see if I could make something integrated into the
// mux subrouters, but unfortunately those freaking suck and are impossible to work with. So in
// the end I simply abstracted the logic so I could have a function that would work for http.handlerfun
// and http.handler and will leave the actual implementation details to some other package.
//
// Why am I telling you this? I just needed to bitch about how I spent my Saturday morning.
func ProtectedHandler(login Login, next http.Handler) http.Handler {
	return ProtectedFunc(login, next.ServeHTTP)
}
