package database

import (
	"database/sql"
	"fmt"

	"github.com/cj-dimaggio/LinkLetter/config"
	"github.com/cj-dimaggio/LinkLetter/logger"
	_ "github.com/lib/pq" // Makes the postgres driver available
)

func configToDatabaseParams(conf config.Config) string {
	sslOption := "disable"
	if conf.SQLUseSSL {
		sslOption = "require"
	}

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		conf.SQLUser, conf.SQLPassword, conf.SQLDB, conf.SQLHost, conf.SQLPort, sslOption)
}

// ConnectToDB opens up a connection to the database and returns
// a reference to it.
func ConnectToDB(conf config.Config) *sql.DB {
	db, err := sql.Open("postgres", configToDatabaseParams(conf))
	if err != nil {
		logger.Error.Printf("Could not create a connection to the postgres database")
		panic(err)
	}
	return db
}
