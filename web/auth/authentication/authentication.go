package authentication

import (
	"net/http"

	"fmt"

	"github.com/Ssawa/LinkLetter/logger"
	"github.com/gorilla/sessions"
)

const (
	loginPage string = "/login"
)

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
	session, err := cookies.Get(req, "session")
	if err != nil {
		logger.Error.Printf("Encountered error while getting session: %s", err)
		return false, err
	}

	val := session.Values["isAuthenticated"]
	if authenticated, ok := val.(bool); ok {
		return authenticated, nil
	}

	return false, nil
}

// AuthProtected is an http middleware that checks if a user is authenticated. If the user
// isn't then he/she is redirected to /login, if the user is, then the request continues
// normally
func AuthProtected(cookies *sessions.CookieStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// I was originally thinking about making this middleware route bases (so you would
		// have to explicitly set it for every route you wanted behind authentication) but
		// quickly realized that, for our specific application, every page except for the login
		// would be under authentication. While adding it for every route might be more explicit,
		// the repetitive tedium of it would also increase the chance of error (someone accidentally
		// leaving it out), so let's just make it a piece of middleware that can work on the
		// entire router at once just to be safe.
		if r.URL.Path != loginPage {
			auth, err := IsAuthenticated(r, cookies)

			if err != nil {
				logger.Error.Printf("Encountered error while doing auth middleware: %s", err)
				http.Error(w, fmt.Sprintf("Encountered error: %s", err), 500)
				return
			}

			if !auth {
				// I kind of think this should be a 303 ("See Other") rather than
				// a 302 ("Found"), but after doing some research it looks like 302s
				// are the standard for cases like this (I believe it's what google uses)
				// so we'll just go with that.
				http.Redirect(w, r, loginPage, 302)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
