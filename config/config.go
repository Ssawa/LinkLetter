// Package config is responsible for defining and gathering configuration data
// for use in the application
package config

// It's probably worth noting that there's probably no reason we should be writing
// our own config system. I'm sure there's plenty of great libraries out there that
// do everything we need. But this project is about learning and trying to learn
// the best way of doing things. Besides, the ones I came across were by people on
// GitHub with  ridiculous user names like "codekilla". Am I supposed to put
// "codekilla" in my imports? Come on.

import (
	"flag"
	"os"
	"strconv"
)

// Config contains all the configurable data needed to run the
// application.
type Config struct {
	// I committed myself to making this a defined struct because I liked the
	// idea of having the compiler checking for spelling typos and IDE code
	// completion. Looking at it now though, I think it should maybe just be
	// switched to a map with string keys. It would make iterating over configs
	// flags much easier, and not require messy introspection. That alone
	// could make the process of adding new config options easier.

	WebPort              int
	SQLPort              int
	SQLHost              string
	SQLDB                string
	SQLUser              string
	SQLPassword          string
	SQLUseSSL            bool
	URLBase              string
	SecretKey            string
	AuthorizationPattern string
	GoogleClientID       string
	GoogleClientSecret   string
}

// GetEnvStringDefault wraps os.Getenv to get an environment variable as a
// string and supporting a default option.
func GetEnvStringDefault(env string, defaultValue string) string {
	value := os.Getenv(env)
	if value == "" {
		value = defaultValue
	}
	return value
}

// GetEnvIntDefault gets an environment variable as an int and supports
// a default option.
func GetEnvIntDefault(env string, defaultInt int) int {
	i, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		i = defaultInt
	}
	return i
}

// ParseForConfig grabs required information from the program args
// and environment variables and creates a Config object. Program
// arguments take precedence over environment variables.
func ParseForConfig() Config {
	// I'm not happy with this. What we're essentially doing is defining our configuration
	// in two separate places, right? The structural make up in the struct and the defaults
	// and environment data here. And while the argument can be made that that's a fair
	// decoupalation, it just doesn't *feel* right to me. This whole thing might need to
	// be rethought.
	//
	// Also, this does not currently support config files, only allowing configs from program
	// args or environment variables, and that needs to be added in sooner or later.

	conf := Config{

		// Okay, so I initially wanted it so that all environment variables would be prefaced
		// with "LINKLETTER_" so that they could play nicely with other programs. But, as it
		// turns out, Heroku insists on passing the binding port in under the environment
		// variable PORT, and we're using Heroku so here we are. Now does that mean that
		// we should go ahead, admit defeat, and just shear off "LINKLETTER_" from the rest
		// of env vars? Yeah, would probably reduce confusion. I just don't have the heart to
		// do it.
		WebPort:              GetEnvIntDefault("PORT", 8080),
		SQLPort:              GetEnvIntDefault("LINKLETTER_SQLPORT", 9753),
		SQLHost:              GetEnvStringDefault("LINKLETTER_SQLHOST", "127.0.0.1"),
		SQLDB:                GetEnvStringDefault("LINKLETTER_SQLDB", "linkletter"),
		SQLUser:              GetEnvStringDefault("LINKLETTER_SQLUSER", "linkletter"),
		SQLPassword:          GetEnvStringDefault("LINKLETTER_SQLPASSWORD", "pass"),
		SecretKey:            GetEnvStringDefault("LINKLETTER_SECRETKEY", "secret123"),
		URLBase:              GetEnvStringDefault("LINKLETTER_URLBASE", "http://localhost:8080"),
		AuthorizationPattern: GetEnvStringDefault("LINKLETTER_AUTHORIZATIONPATTERN", "localprojects\\.(com|net)"),
		GoogleClientID:       GetEnvStringDefault("LINKLETTER_GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:   GetEnvStringDefault("LINKLETTER_GOOGLE_CLIENT_SECRET", ""),
	}

	flag.IntVar(&conf.WebPort, "webPort", conf.WebPort, "The port to run the web application on")
	flag.IntVar(&conf.SQLPort, "sqlPort", conf.SQLPort, "The port the SQL server is running on")
	flag.StringVar(&conf.SQLHost, "sqlHost", conf.SQLHost, "The SQL server host")
	flag.StringVar(&conf.SQLDB, "sqlDB", conf.SQLDB, "The SQL database name to connect to")
	flag.StringVar(&conf.SQLUser, "sqlUser", conf.SQLUser, "The username to use when connecting to SQL database")
	flag.StringVar(&conf.SQLPassword, "sqlPassword", conf.SQLPassword, "The password to use when conneting to SQL database")
	flag.BoolVar(&conf.SQLUseSSL, "sqlUseSSL", false, "Whether or not SQL connection should be over SSL")
	flag.StringVar(&conf.SecretKey, "secretKey", conf.SecretKey, "The secret key to use to sign cookies")
	flag.StringVar(&conf.URLBase, "urlBase", conf.URLBase, "The base URL path the webserver will be hosted at (will be used for OAuth2 redirect url generation)")
	flag.StringVar(&conf.AuthorizationPattern, "authorizationPattern", conf.AuthorizationPattern, "The regex pattern to match against hosted domains for authorization")
	flag.StringVar(&conf.GoogleClientID, "googleClientID", conf.GoogleClientID, "Google OAuth2 client ID")
	flag.StringVar(&conf.GoogleClientSecret, "googleClientSecret", conf.GoogleClientSecret, "Google OAuth2 client secret")

	flag.Parse()
	return conf
}
