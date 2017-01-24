// Package logger is a simple interface for keeping people informed about what the application is doing
package logger

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// I've been firmly indoctrinated in the current understanding that global variables are evil. I don't
// actually have many thoughts on the subject, but I've heard it enough times so that whenever I
// do write a global variable I feel a deep sense of shame and uncomfortability. As such, I try
// to avoid them if I can. Logging, however, is somewhere I'm willing to make an exception. It's
// important, when developing, to be able to just import an object and log what you need to log.
// If you force someone to fart around with things like inversion of control and dependency
// injection just to satisfy programming dogma, the end result is that person will probably just
// not use the logger. Which is the worst result. It's also, generally, easy and benign enough
// to not make writing tests a chore.
var (
	// Debug logs verbose messages for use in debugging
	Debug = log.New(ioutil.Discard, "", 0)

	// Info logs regular messages to keep track of status of application
	Info = log.New(ioutil.Discard, "", 0)

	// Warning logs messages that might lead to disrupt behavior of the application
	Warning = log.New(ioutil.Discard, "", 0)

	// Error logs messages when things go very wrong and will screw up the application
	Error = log.New(ioutil.Discard, "", 0)
)

// InitLogging initializes the global logging variables for use.
func InitLogging(debugOutput io.Writer, infoOutput io.Writer, warningOutput io.Writer, errorOutput io.Writer, logFormat int) {
	Debug = log.New(debugOutput, "DEBUG: ", logFormat)
	Info = log.New(infoOutput, "INFO: ", logFormat)
	Warning = log.New(warningOutput, "WARNING: ", logFormat)
	Error = log.New(errorOutput, "ERROR: ", logFormat)
}

// InitLoggingDefault initializes logging with the default settings of logging everything to standard out.
func InitLoggingDefault() {
	InitLogging(os.Stdout, os.Stdout, os.Stdout, os.Stdout, log.Ldate|log.Ltime|log.Lshortfile)
}

// SetLogLevel sets the lowest log level of which to show and suppresses all those lower.
func SetLogLevel(level string) {
	//This guy needs a lot of work. For one, it's not very DRY, I imagine a lot of it can be
	// factored out into functions. More so, it's not idemportent. If you call SetLogLevel(WARNING)
	// and then SetLogLevel(INFO), info still remains pointed to Discard.

	switch strings.ToLower(level) {
	case "info":
		Debug.SetOutput(ioutil.Discard)
	case "warning":
		Debug.SetOutput(ioutil.Discard)
		Info.SetOutput(ioutil.Discard)
	case "error":
		Debug.SetOutput(ioutil.Discard)
		Info.SetOutput(ioutil.Discard)
		Warning.SetOutput(ioutil.Discard)
	}
}

// LogHTTPRequests wraps an http handler to log every request that comes in.
func LogHTTPRequests(logger *log.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
