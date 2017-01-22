package main

import (
	"fmt"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/database"
	"github.com/Ssawa/LinkLetter/logger"
	"github.com/Ssawa/LinkLetter/web"
)

func main() {
	logger.InitLoggingDefault()
	logger.Debug.Printf("Determining configs...")
	conf := config.ParseForConfig()

	db := database.ConnectToDB(conf)
	database.DoMigrations(db)

	logger.Debug.Println("Creating server...")
	server := web.CreateServer(conf, db)
	logger.Info.Printf("Starting server...")
	http.ListenAndServe(fmt.Sprintf(":%d", conf.WebPort), logger.LogHTTPRequests(logger.Info, server.Router))
}
