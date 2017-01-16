package web

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateServer(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("../")

	server := CreateServer(config.Config{
		SecretKey: "test",
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	server.Router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}
