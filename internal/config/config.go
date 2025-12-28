package config

import (
	"log"
	"os"

	"github.com/ItsMonish/StartPage/internal/types"
	"gopkg.in/yaml.v3"
)

func GetConfig(logger *log.Logger, configPath string) types.RootConfiguration {
	configFile, err := os.Open(configPath)
	if err != nil {
		logger.Println("Cannot open configuration file at " + configPath)
		os.Exit(1)
	}
	defer configFile.Close()
	logger.Println("Reading from config file at " + configPath)

	var config types.RootConfiguration

	yamlDecoder := yaml.NewDecoder(configFile)
	if err := yamlDecoder.Decode(&config); err != nil {
		logger.Println("Failed to deserialize the config file")
		logger.Println(err.Error())
		os.Exit(1)
	}

	if config.Props.RetryInterval == 0 {
		config.Props.RetryInterval = 5
	}
	if config.Props.RefreshInterval == 0 {
		config.Props.RefreshInterval = 60
	}

	return config
}
