package main

import (
	"github.com/joho/godotenv"

	"github.com/fsvxavier/default-vertical-slice/cmd/webserver"
	logger "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

// @title		template API
// @version	1.0
// @host		localhost:8086
// @schemes	http.
func main() {
	// To load our environmental variables.
	err := godotenv.Load(".env")
	if err != nil {
		logger.DebugOutCtx("\nNo .env file avaliable seaching ENVIROMENTS system")
	}

	webserver.Run()
}
