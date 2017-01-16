package main

import (
	"fmt"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/logger"
	"github.com/Ssawa/LinkLetter/web"
)

func main() {
	logger.InitLoggingDefault()
	logger.Debug.Printf("Determining configs...")
	conf := config.ParseForConfig()

	logger.Debug.Println("Creating server...")
	server := web.CreateServer(conf)
	logger.Info.Printf("Starting server...")
	http.ListenAndServe(fmt.Sprintf(":%d", conf.WebPort), logger.LogHTTPRequests(logger.Info, server.Router))
}
