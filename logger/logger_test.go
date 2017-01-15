package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	lastWritten string
)

type TestingWriter struct {
}

// Write doesn't take a pointer argument so you can't set data to a field here.
// Instead, we'll just use a global variable
func (writer TestingWriter) Write(p []byte) (n int, err error) {
	lastWritten = string(p)
	return len(p), nil
}

func TestInitLogging(t *testing.T) {
	debugWritter := TestingWriter{}
	infoWritter := TestingWriter{}
	warningWritter := TestingWriter{}
	errorWritter := TestingWriter{}

	lastWritten = ""

	InitLogging(debugWritter, infoWritter, warningWritter, errorWritter, 0)
	Debug.Println("THIS IS A DEBUG")
	assert.Equal(t, "DEBUG: THIS IS A DEBUG\n", lastWritten)

	Info.Println("THIS IS AN INFO")
	assert.Equal(t, "INFO: THIS IS AN INFO\n", lastWritten)

	Info.Println("THIS IS A WARNING")
	assert.Equal(t, "INFO: THIS IS A WARNING\n", lastWritten)

	Info.Println("THIS IS AN ERROR")
	assert.Equal(t, "INFO: THIS IS AN ERROR\n", lastWritten)
}

func TestSetLogLevel(t *testing.T) {
	debugWritter := TestingWriter{}
	infoWritter := TestingWriter{}
	warningWritter := TestingWriter{}
	errorWritter := TestingWriter{}

	lastWritten = ""
	InitLogging(debugWritter, infoWritter, warningWritter, errorWritter, 0)

	SetLogLevel("INFO")
	Debug.Println("SUPRESSED?")
	assert.Equal(t, "", lastWritten)

	SetLogLevel("WARNING")
	Info.Println("SUPRESSED?")
	assert.Equal(t, "", lastWritten)

	SetLogLevel("ERROR")
	Warning.Println("SUPRESSED?")
	assert.Equal(t, "", lastWritten)
}
