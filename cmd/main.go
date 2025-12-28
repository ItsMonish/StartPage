package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/server"
)

var (
	configPath string
	port       int
	logging    bool
	logger     *log.Logger
)

func main() {
	userConfig, _ := os.UserConfigDir()
	defaultConfig := userConfig + "/startpage/config.yml"
	logFilePath := userConfig + "/startpage/application.log"

	flag.StringVar(&configPath, "config", defaultConfig, "Path to the config file")
	flag.IntVar(&port, "port", 8080, "Port to open the server on")
	flag.BoolVar(&logging, "log", false, "Redirect log to STDOUT")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if !logging {
		logFile, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		wrt := io.Writer(logFile)
		log.SetOutput(wrt)
		logger = log.New(wrt, "", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	}

	config := config.GetConfig(logger, configPath)

	server.StartServer(logger, config)

}
