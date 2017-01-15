package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestingWriter struct {
	lastWritten *string
}

func (writer TestingWriter) Write(p []byte) (n int, err error) {
	*writer.lastWritten = string(p)
	return len(p), nil
}

func TestInitLogging(t *testing.T) {
	debugWritter := TestingWriter{new(string)}
	infoWritter := TestingWriter{new(string)}
	warningWritter := TestingWriter{new(string)}
	errorWritter := TestingWriter{new(string)}

	InitLogging(debugWritter, infoWritter, warningWritter, errorWritter, 0)
	Debug.Println("THIS IS A DEBUG")
	assert.Equal(t, "DEBUG: THIS IS A DEBUG\n", *debugWritter.lastWritten)

	Info.Println("THIS IS AN INFO")
	assert.Equal(t, "INFO: THIS IS AN INFO\n", *infoWritter.lastWritten)

	Warning.Println("THIS IS A WARNING")
	assert.Equal(t, "WARNING: THIS IS A WARNING\n", *warningWritter.lastWritten)

	Error.Println("THIS IS AN ERROR")
	assert.Equal(t, "ERROR: THIS IS AN ERROR\n", *errorWritter.lastWritten)
}

func TestSetLogLevel(t *testing.T) {
	debugWritter := TestingWriter{new(string)}
	infoWritter := TestingWriter{new(string)}
	warningWritter := TestingWriter{new(string)}
	errorWritter := TestingWriter{new(string)}

	InitLogging(debugWritter, infoWritter, warningWritter, errorWritter, 0)

	SetLogLevel("INFO")
	Debug.Println("SUPRESSED?")
	assert.Equal(t, "", *debugWritter.lastWritten)

	SetLogLevel("WARNING")
	Info.Println("SUPRESSED?")
	assert.Equal(t, "", *infoWritter.lastWritten)

	SetLogLevel("ERROR")
	Warning.Println("SUPRESSED?")
	assert.Equal(t, "", *warningWritter.lastWritten)
}
