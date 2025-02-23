package config

import (
	"log"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Port  int                       `yaml:"port"`
	Rss   map[string][]TitleURLItem `yaml:"rss"`
	Yt    []TitleURLItem            `yaml:"youtube"`
	Links QuickLinks                `yaml:"quicklinks"`
}

type TitleURLItem struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}

type QuickLinks struct {
	List1Name string         `yaml:"list1Name"`
	List2Name string         `yaml:"list2Name"`
	List3Name string         `yaml:"list3Name"`
	List4Name string         `yaml:"list4Name"`
	List1     []TitleURLItem `yaml:"list1"`
	List2     []TitleURLItem `yaml:"list2"`
	List3     []TitleURLItem `yaml:"list3"`
	List4     []TitleURLItem `yaml:"list4"`
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

	//logger.Println("List 1: ", config.Links.List1Name)
	//for _, item := range config.Links.List1 {
	//	logger.Println("Title: ", item.Title, " - ", item.Url)
	//}
	//logger.Println("List 2: ", config.Links.List2Name)
	//for _, item := range config.Links.List2 {
	//	logger.Println("Title: ", item.Title, " - ", item.Url)
	//}
	//logger.Println("List 3: ", config.Links.List3Name)
	//for _, item := range config.Links.List3 {
	//	logger.Println("Title: ", item.Title, " - ", item.Url)
	//}
	//logger.Println("List 4: ", config.Links.List4Name)
	//for _, item := range config.Links.List4 {
	//	logger.Println("Title: ", item.Title, " - ", item.Url)
	//}

	return config
}
