package database

import (
	"database/sql"
	"fmt"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/logger"
	_ "github.com/lib/pq" // Makes the postgres driver available
)

// ConnectToDB opens up a connection to the database and returns
// a pointer to the connection
func ConnectToDB(conf config.Config) *sql.DB {
	databaseParams := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d",
		conf.SqlUser, conf.SqlPassword, conf.SqlDB, conf.SqlHost, conf.SqlPort)
	db, err := sql.Open("postgres", databaseParams)
	if err != nil {
		logger.Error.Printf("Could not create a connection to the postgres database")
		panic(err)
	}
	return db
}
