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
	port       int
	configPath string
	logging    bool
	logger     *log.Logger
)

func main() {
	userConfig, _ := os.UserConfigDir()
	defaultConfig := userConfig + "/startpage/config.yml"
	logFilePath := userConfig + "/startpage/application.log"

	flag.StringVar(&configPath, "config", defaultConfig, "Path to config file")
	flag.IntVar(&port, "port", 8080, "Port to open server on")
	flag.BoolVar(&logging, "log", false, "Output log to STDOUT")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if logging {
		logFile, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		wrt := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(wrt)
		logger = log.New(wrt, "", log.LUTC|log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "", log.LUTC|log.LstdFlags|log.Lshortfile)
	}

	conf := config.GetConfig(logger, configPath)

	//Command line value takes precedence over config value
	conf.Props.Port = port

	server.StartServer(logger, conf)
}
