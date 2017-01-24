package database

import (
	"database/sql"
	"fmt"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/logger"
	_ "github.com/lib/pq" // Makes the postgres driver available
)

// ConnectToDB opens up a connection to the database and returns
// a pointer to it.
func ConnectToDB(conf config.Config) *sql.DB {
	sslOption := "disable"
	if conf.SQLUseSSL {
		sslOption = "require"
	}

	databaseParams := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		conf.SQLUser, conf.SQLPassword, conf.SQLDB, conf.SQLHost, conf.SQLPort, sslOption)
	db, err := sql.Open("postgres", databaseParams)
	if err != nil {
		logger.Error.Printf("Could not create a connection to the postgres database")
		panic(err)
	}
	return db
}
