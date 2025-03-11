package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Props Properties                `yaml:"properties"`
	Rss   map[string][]TitleURLItem `yaml:"rss"`
	Yt    []TitleURLItem            `yaml:"youtube"`
	Links QuickLinks                `yaml:"quicklinks"`
}

type Properties struct {
	Port            int `yaml:"port"`
	RefreshInterval int `yaml:"refreshInterval"`
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
		os.Exit(1)
	}

	defer configFile.Close()

	var config Configuration

	logger.Println("Reading from config file: " + configPath)

	yamlDecoder := yaml.NewDecoder(configFile)
	if err := yamlDecoder.Decode(&config); err != nil {
		logger.Fatal("Error reading the config file")
		logger.Fatal("Check syntax of the file")
		os.Exit(1)
	}

	return config
}
