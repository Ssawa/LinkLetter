package database

import "testing"
import "github.com/Ssawa/LinkLetter/config"
import "github.com/stretchr/testify/assert"

func TestConfigToDatabaseParams(t *testing.T) {
	conf := config.Config{
		SQLUseSSL:   true,
		SQLUser:     "testUser",
		SQLPassword: "testPassword",
		SQLDB:       "testDB",
		SQLHost:     "testHost",
		SQLPort:     55,
	}

	assert.Equal(t, "user=testUser password=testPassword dbname=testDB host=testHost port=55 sslmode=require", configToDatabaseParams(conf))

	conf.SQLUseSSL = false

	assert.Equal(t, "user=testUser password=testPassword dbname=testDB host=testHost port=55 sslmode=disable", configToDatabaseParams(conf))
}
