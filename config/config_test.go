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
	os.Setenv("PORT", "7000")
	os.Setenv("LINKLETTER_SQLHOST", "testhost")
	os.Setenv("LINKLETTER_SQLDB", "testdb")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{os.Args[0], "-sqlDB", "override", "-sqlUser", "testuser", "-sqlUseSSL"}
	conf := ParseForConfig()

	assert.Equal(t, 7000, conf.WebPort)
	assert.Equal(t, "testhost", conf.SQLHost)
	assert.Equal(t, "override", conf.SQLDB)
	assert.Equal(t, "testuser", conf.SQLUser)
	assert.Equal(t, true, conf.SQLUseSSL)
}
