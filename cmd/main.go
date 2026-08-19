package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/database"
	"github.com/ItsMonish/StartPage/internal/server"
)

var (
	configPath string
	dbPath     string
	port       int
	logging    bool
	logger     *log.Logger
)

func main() {
	userConfig, _ := os.UserConfigDir()
	logFilePath := userConfig + "/startpage/application.log"

	flag.StringVar(&configPath, "config", "", "Path to the config file (default \"$HOME/.config/startpage/config.yml\")")
	flag.StringVar(&dbPath, "db", "", "Path to the database file (default \"$HOME/.config/startpage/database.db\")")
	flag.IntVar(&port, "port", 8080, "Port to open the server on")
	flag.BoolVar(&logging, "log", false, "Redirect log to STDOUT")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if configPath == "" {
		configPath = userConfig + "/startpage/config.yml"
	}

	if dbPath != "" {
		dbPath = userConfig + "/startpage/database.db"
	}

	if !logging {
		logFile, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		wrt := io.Writer(logFile)
		log.SetOutput(wrt)
		logger = log.New(wrt, "", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	}

	config := config.GetConfig(logger, configPath)
	config.Props.Port = port
	config.Props.DatabasePath = dbPath

	err := database.InitDb(config.Props.DatabasePath)
	if err != nil {
		logger.Println("Error in initializing the database")
		logger.Println(err.Error())
		os.Exit(1)
	}

	server.StartServer(logger, config)

}
