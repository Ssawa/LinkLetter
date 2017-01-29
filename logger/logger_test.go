package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogging(t *testing.T) {
	log := CreateDummyLogger()

	Debug.Println("THIS IS A DEBUG")
	assert.Equal(t, "DEBUG: THIS IS A DEBUG\n", log.Debug.Last())

	Info.Println("THIS IS AN INFO")
	assert.Equal(t, "INFO: THIS IS AN INFO\n", log.Info.Last())

	Warning.Println("THIS IS A WARNING")
	assert.Equal(t, "WARNING: THIS IS A WARNING\n", log.Warning.Last())

	Error.Println("THIS IS AN ERROR")
	assert.Equal(t, "ERROR: THIS IS AN ERROR\n", log.Error.Last())
}

func TestSetLogLevel(t *testing.T) {
	log := CreateDummyLogger()

	SetLogLevel("INFO")
	Debug.Println("SUPPRESSED?")
	assert.Equal(t, "", log.Debug.Last())

	SetLogLevel("WARNING")
	Info.Println("SUPPRESSED?")
	assert.Equal(t, "", log.Info.Last())

	SetLogLevel("ERROR")
	Warning.Println("SUPPRESSED?")
	assert.Equal(t, "", log.Warning.Last())
}

func TestLogHTTPRequests(t *testing.T) {
	log := CreateDummyLogger()

	mux := http.NewServeMux()
	handler := LogHTTPRequests(Debug, mux)

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Contains(t, log.Debug.Last(), "GET /test")
}
