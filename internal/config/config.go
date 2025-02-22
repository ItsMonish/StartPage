package config

import (
	"log"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Port int                  `yaml:"port"`
	Rss  map[string][]RSSFeed `yaml:"rss"`
	Yt   []YTFeed             `yaml:"youtube"`
}

type RSSFeed struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}

type YTFeed struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}

func GetConfig(logger *log.Logger, configPath string) Configuration {
	configFile, err := os.Open(configPath)
	if err != nil {
		logger.Fatal("There was an error reading config file at " + configPath)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}

	defer configFile.Close()

	var config Configuration

	logger.Println("Reading from config file: " + configPath)

	yamlDecoder := yaml.NewDecoder(configFile)
	if err := yamlDecoder.Decode(&config); err != nil {
		logger.Fatal("Error reading the config file")
		logger.Fatal("Check syntax of the file")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}

	//logger.Println("Server port: ", config.Port)
	//for category, feeds := range config.Rss {
	//	logger.Println("Category:", category)
	//	for _, feed := range feeds {
	//		logger.Printf("  - %s: %s\n", feed.Title, feed.Url)
	//	}
	//}

	//for _, channel := range config.Yt {
	//	logger.Println(channel.Title, " - ", channel.Url)
	//}

	return config
}
