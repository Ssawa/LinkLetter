package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	// Debug logs verbose messages for use in debugging
	Debug *log.Logger

	// Info logs regular messages to keep track of status of application
	Info *log.Logger

	// Warning logs messages that might lead to disrupt behavior of the application
	Warning *log.Logger

	// Error logs messages when things go very wrong and will screw up the application
	Error *log.Logger
)

// InitLogging initializes variables for use in logging
func InitLogging(debugOutput io.Writer, infoOutput io.Writer, warningOutput io.Writer, errorOutput io.Writer, logFormat int) {
	Debug = log.New(debugOutput, "DEBUG: ", logFormat)
	Info = log.New(infoOutput, "INFO: ", logFormat)
	Warning = log.New(warningOutput, "WARNING: ", logFormat)
	Error = log.New(errorOutput, "ERROR: ", logFormat)
}

// InitLoggingDefault initializes logging with default settings
func InitLoggingDefault() {
	InitLogging(os.Stdout, os.Stdout, os.Stdout, os.Stdout, log.Ldate|log.Ltime|log.Lshortfile)
}

// SetLogLevel supresses certain log levels to reduce noise (CURRENTLY THIS IS NOT IDEMPOTENT, IT SHOULD ONLY BE CALLED ONCE IN AN APPLICATION)
func SetLogLevel(level string) {
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
