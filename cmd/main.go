package main

import (
	"log"
	"os"

	"github.com/ItsMonish/StartPage/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "", log.LUTC|log.LstdFlags)

	server.StartServer(logger, 8080)
}
