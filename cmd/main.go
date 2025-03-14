package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/server"
)

var (
	port       int
	configPath string
)

func main() {
	logger := log.New(os.Stdout, "", log.LUTC|log.LstdFlags|log.Lshortfile)

	defaultConfig, _ := os.UserConfigDir()
	defaultConfig += "/startpage/config.yml"

	flag.StringVar(&configPath, "config", defaultConfig, "Path to config file")
	flag.IntVar(&port, "port", 8080, "Port to open server on")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	conf := config.GetConfig(logger, configPath)

	//Command line value takes precedence over config value
	conf.Props.Port = port

	server.StartServer(logger, conf)
}
