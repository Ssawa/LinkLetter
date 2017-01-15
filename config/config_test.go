package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvStringDefault(t *testing.T) {
	os.Setenv("TESTENV", "RESULT")

	assert.Equal(t, "RESULT", GetEnvStringDefault("TESTENV", "NOTRESULT"))

	assert.Equal(t, "DEFAULTRESULT", GetEnvStringDefault("MAKEBELIEVEENV1234", "DEFAULTRESULT"))
}

func TestGetEnvIntDefault(t *testing.T) {
	os.Setenv("TESTENV", "10")

	assert.Equal(t, 10, GetEnvIntDefault("TESTENV", 0))

	assert.Equal(t, 5, GetEnvIntDefault("MAKEBELIEVEENV1234", 5))
}

func TestParseForConfig(t *testing.T) {
	os.Setenv("LINKLETTER_WEBPORT", "7000")
	os.Setenv("LINKLETTER_SQLHOST", "testhost")
	os.Setenv("LINKLETTER_SQLDB", "testdb")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{os.Args[0], "-sqlDB", "override", "-sqlUser", "testuser"}
	conf := ParseForConfig()

	assert.Equal(t, 7000, conf.webPort)
	assert.Equal(t, "testhost", conf.sqlHost)
	assert.Equal(t, "override", conf.sqlDB)
	assert.Equal(t, "testuser", conf.sqlUser)
}
