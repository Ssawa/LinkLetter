package oauth2

// OAuth2 is an incredibly simple protocol. In fact, it's been argued that it's
// simplicity comes at the expense of security (an argument this particular programmer
// somewhat agrees with). As such, it would be totally incorrect to think of OAuth2 as
// a replacement for it's much more secure brother OAuth, it is simply another approach.
//
// I find that, like most development concepts, while OAuth2 has firmly entered the
// programmer's lexicon there are precious few resources for what it actually is. The
// first question one might ask is what exactly does OAuth do? Well the "Auth" in OAuth
// stands for "Authorization" (and *only* "authorization", but we'll get to that when
// we talk about our Authentication method). "Authorization", is the processes of
// determining what level of access a user, or in some cases a program, has over a resource.
// OAuth2 attempts to provide a simple way for a user to give his or her consent to an outside
// agent to perform actions on his/her accounts.
//
// An example: Jane is writing a program that will analyze your emails and fill out your calender
// with mentioned events. The question we're concerned with is how does she actually get these
// emails? What would be great is if she could use one of gmail's APIs to download all her user's
// emails. But how would she do this? A naive approach might be to ask for the user's gmail
// password, I mean if you have a user's password you can do anything you want. And that's precisley
// why that approach is naive; if a user were to give Jane their gmail password Jane would suddenly
// have access to the user's *entire* account. Jane could download the emails but she could also
// delete them, send new ones, even change the user's password locking them out. But even if we
// could trust Jane, we're now depending on her to store our password securely, and if she loses
// it we lose our gmail account. And forget about password storage, what about password transmission?
// What if Jane isn't using HTTPS and that password gets sent over the internet in completely plain text?
// You can see why authorization is a tricky field. What Jane needs is a way for the user to allow her to
// read, and ONLY read, their emails and nothing else and not have the user depend on her having to
// securely handle their sensitive information. And that's the role that OAuth fills.
//
// So how does it do this? To hear people talk about it you might start to be led to believe that
// OAuth is some kind of library or program itself that handles this functionality but that
// is not the case at all. OAuth is simply a protocol; but not a protocol in the sense that
// TCP and HTTP are considered protocols. There is no "OAuth" data format, no "OAuth" parsers,
// and no libraries you can download and type in "start_oauth_server()". OAuth is a protocol
// only in the sense that it is a convention which several APIs choose to follow. It's a fairly
// loose standard, to be honest, and every provider (google, github, linkedin, etc) implements
// it slightly differently.
//
// In some sense, OAuth2 does nothing more than suggest how to safely ping pong http messages
// in a variety of scenerios. These "scenerios" are sometimes referred to as "flows". For example,
// did you know that on a web browser the part after a hash ("#") character in a url (as in:
// https://en.wikipedia.org/wiki/article#section) is referred to as a "fragment" and that this
// part of the URL is never actually sent in the request to the server? Well the creators of
// OAuth2 did, and so they developed a "flow" in which a single page web application can perform
// secure authorization without ever having to transmit sensitive data to potentially insecure
// servers.
//
// We're not creating a single page web application, however, so we'll be using a different flow.
// Instead we'll be leveraging our backend webserver to handle this OAuth2 negotiation, rather
// than putting the logic into client facing code.

import (
	"net/http"
	"regexp"

	"github.com/cj-dimaggio/LinkLetter/logger"
	"github.com/cj-dimaggio/LinkLetter/web/auth/authentication"
	"github.com/gorilla/sessions"
)

// OAuth2 handles interaction with OAuth2 servers in a generalized fashion.
//
// As mentioned above, OAuth2 is a pretty loose standard and every implementation is slightly
// different and that might make generalizing the process into an interface a bit of an exercise
// in futility but it might prove to be a good opportunity to teach the general "flow" we'll be
// using to perform authentication over OAuth2.
//
// The general convention we'll be following is that those parameters that are explicitly defined
// in the OAuth2 standard are denoted by the function signature and passed in (such as clientID
// and redirectURL). While those that are up to provider implementation (such as a refresh token)
// are expected to be handled by the implementing class internally.
type OAuth2 interface {

	// GenerateAuthorizationURL is the first piece in the puzzle and there is really nothing
	// magical here. All this does is create a url to an OAuth2 provider (such as google)
	// with a number of query parameters set. Some of these query parameters are dictated by
	// the OAuth2 specification, such as redirectURL, clientID, and scope while others are
	// specific to the provider. This URL will generally be placed on some link or button that says
	// something along the lines of "Sign in with Google" and when clicked will take the user to the
	// provider's website where they can sign in with the provider and authorize your application.
	// So what do "redirectURL", "clientID", and "scope" do?
	//
	// "Scope" is the simplest to explain, this is simply a string that tells the provider what
	// level of access your application is asking for. If you've ever signed in with google before
	// you will have seen one of those pages where google says "This application is asking for permission
	// to view your: email information etc..." "Scope" forces us to tell the user exactly what we're
	// planning on doing with their account and restricts our API access to only those actions. Every
	// provider has their own list of valid scopes so it is very much implementation specific, but you
	// can generally depend on it being a required value.
	//
	// "Client ID" is a bit more vague. Before your application can integrate with an OAuth2 provider
	// you'll generally need to go to the prover's developer section, register your account, and
	// generate a Client ID and a Client Secret. These two pieces of information are simply used to
	// identify and validate your application on the providers system so that another application cannot
	// impersonate yours or, if your application is acting irresponsibly, the provider can easily pull
	// your access. The Client Secret is *only* ever to be known by your application, while Client ID
	// can be safely passed around publicly (in this case we're passing it to our user). Think of it
	// like a public key and a private key.
	//
	// "Redirect URL" comes up a couple of times in OAuth2 and might seem odd at first but I assure you
	// is as simple as it sounds. Redirect URL simply defines where the provider should redirect the user
	// after he/she as accepted or declined authorization. The url that the provider directs the user
	// to will also contain some information in the query parameters, such as the authorization code
	// if the user accepted or an error if something went wrong. This redirect url will usually need to
	// be explicitly entered into the provider's developer section where you generated your Client ID
	// and Secret, otherwise a nefarious agent could take your publically accessible Client ID and redirect
	// a user to a potentially malicious or inappropriate site.
	GenerateAuthorizationURL(redirectURL, clientID, scope string) string

	// ExtractAuthorizationCode is to be called in the context of the redirectURL that was passed into the
	// GenerateAuthorizationURL function. When a user is redirected to the "redirect url" the provider will
	// include the authorization code (or potentially errors) into query strings of the url, and it is the
	// purpose of this function to extract that authorization code from the user's request.
	ExtractAuthorizationCode(req *http.Request) (string, error)

	// GenerateAccessTokenURL takes the authorization code we should have just extracted with ExtractAuthorizationCode
	// and generates a request to the provider's APIs for getting an access token from that authorization code.
	// The access token is what we'll actually be using to authenticate ourselves with the provider's APIs when
	// we want to interact with them on behalf of the user. To get this we pass in not only our authorization code
	// but also our client secret, validating that we are who we say we are.
	//
	// I'm going to be honest, I don't know why we pass in our redirect url here, it doesn't get used again like
	// like it did with GenerateAuthorizationURL so I'm assuming its a potential authentication thing? Some providers
	// require it and some don't so let's go ahead and add it, if you don't need it just make it an empty string.
	//
	// Now you may be wondering why OAuth2 requires this extra step of converting an "authorization code" to an
	// "access token". Why not just give us the access token directly in our redirect url parameters rather than
	//  make us send this secondary request. The important thing to keep in mind is that the *entire* basis of OAuth2's
	// security comes from the preassumption that the provider is under TLS while the application may or may not be.
	// Essentially, information can only be assumed to be transmitted securely if it's going to/from the provider's
	// servers, not from yours.
	//
	// Assume the provider *had* given us this access token (which allows us to make requests on behalf of a user) directly
	// through the redirect url. What would that transaction look like? First the user would have made a GET to the
	// provider's server's under HTTPS, then the provider would have sent the user a redirect with this sensitive information
	// also under HTTPS. The user's browser would have than followed the redirect with sensitive information to our server
	// which *may or may not be served under https*. And there in lays our information leak. And not only that, we'd then lose
	// all ability to perform the client id/secret validation to prove we're who we say we are. We can't pass this to the user
	// for his or her initial GET request or else anybody who visits our site would be able to impersonate our application.
	GenerateAccessTokenRequest(authorizationCode, redirectURI, clientID, clientSecret string) *http.Request

	// ExtractAccessToken, much like ExtractAuthorizationCode, simply extracts the access token from the response to our
	// GenerateAccessTokenRequest request so that we can use it for subsequent API calls.
	ExtractAccessToken(resp *http.Response) (string, error)

	// Alright here we are, Authenticate. There are two sides of user management; "Authorization", which as we've discussed,
	// is a system of determining what a user is allowed to do, and "Authentication", validating that the user is actually
	// who they say they are. We've seen how OAuth handles authorization, how does it handle authentication? The answer to that
	// is simple: it doesn't. OAuth and OAuth2 was never meant to simulate a single-sign-on system, it was never meant to
	// authenticate users over google or slack or github. It's the shameful secret of the web world; it's all a horrible hack.
	//
	// I imagine it might have something to do  with the unfortunate choice of name that people assumed that because O*Auth* was
	// so great for *Auth*orization it must naturally be great for *Auth*entication. But it's really just not. Of course, since
	// people have started abusing it this way there have been some efforts to add things to the protocol to support it more
	// cleanly, such as OpenID Connect; but in my mind these are just band-aids that do little more than add some defense against
	// CSRF.
	//
	// The general analogy is that performing authentication with OAuth is like asking someone for the keys to their house
	// to validate that they live there.
	//
	// So if it's so horrible, why are we using it? Well...because it really does just make things easier. Okay, I'm willing to
	// admit I'm part of the problem, but if it's the way Google reccommends you perform authentication with their system and
	// it means I don't have to write my own user management system, I'm willing to go with it.
	//
	// So how are we going to make this work? Well we've successfully retrieved our access token that, hopefully, allows us
	// to query the provider for the user's account information. We then validate that that user is allowed to access our system.
	// Sometimes this is done using the user's account ID but this requires knowing which ones you're willing to accept before hand.
	// Instead, our process is going to be very Google specific, if their "hosted domain" matches our pattern
	// (localprojects.com/localprojects.net for example) then their allowed to use the site. How would that translate to something
	// like allowing logins from someplace like github where they don't validate email addresses? I don't freaking know, man, but
	// I wasted so much time trying to think of how to generalize this that I just wasn't doing anything so I'm putting those problems
	// off until the future.
	Authenticate(accessToken string, pattern *regexp.Regexp) (bool, error)
}

// OAuth2Login implements the OAuth2 login logic using the OAuth2Provider of the user's choice
type OAuth2Login struct {
	ClientID             string
	ClientSecret         string
	AuthorizationPattern *regexp.Regexp
	RedirectURL          string
	Scope                string
	OAuth2Provider       OAuth2
	Cookies              *sessions.CookieStore
}

// ShouldAuthenticate looks at the client id and client secret to determine if we should attempt
// authentication
func (login OAuth2Login) ShouldAuthenticate() bool {
	return login.ClientID != "" && login.ClientSecret != ""
}

// GetCookies returns the stored CookieStore
func (login OAuth2Login) GetCookies() *sessions.CookieStore {
	return login.Cookies
}

// GetAuthorizationURL passes in the necessary parameters to the oauth2provider to generate an authorization url
func (login OAuth2Login) GetAuthorizationURL() string {
	return login.OAuth2Provider.GenerateAuthorizationURL(login.RedirectURL, login.ClientID, login.Scope)
}

// AuthorizationCallbackHandler handles the response to our RedirectURL to perform the OAuth2 logic of retrieving
// the authorization code, access token, and finally determining if the client is actually authorized for our
// application.
func (login OAuth2Login) AuthorizationCallbackHandler(w http.ResponseWriter, req *http.Request) {
	authCode, err := login.OAuth2Provider.ExtractAuthorizationCode(req)
	if err != nil {
		logger.Error.Printf("Was unable to get authorization code for login: %s", err)
		http.Error(w, "Was unable to log you into the system", 500)
		return
	}

	tokenReq := login.OAuth2Provider.GenerateAccessTokenRequest(authCode, login.RedirectURL, login.ClientID, login.ClientSecret)
	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		logger.Error.Printf("Was unable to get access token code for login: %s", err)
		http.Error(w, "Was unable to log you into the system", 500)
		return
	}

	token, err := login.OAuth2Provider.ExtractAccessToken(tokenResp)
	if err != nil {
		logger.Error.Printf("Was unable to extract access token code for login: %s", err)
		http.Error(w, "Was unable to log you into the system", 500)
		return
	}

	authenticated, err := login.OAuth2Provider.Authenticate(token, login.AuthorizationPattern)
	if err != nil {
		logger.Error.Printf("Error occurred while authenticating: %s", err)
		http.Error(w, "An error occurred while trying to authenticate you", 500)
		return
	}

	if !authenticated {
		http.Error(w, "Unfortunately you are not allowed to access this site", 403)
		return
	}

	// Redirect the, now authenticated, user back to the index page
	authentication.LogInUser(login.Cookies, req, w)
	http.Redirect(w, req, "/", 302)
}
