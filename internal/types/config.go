package types

type RootConfiguration struct {
	Props ConfigProperties                `yaml:"properties"`
	Rss   map[string][]ConfigTitleURLItem `yaml:"rss"`
	Yt    []ConfigTitleURLItem            `yaml:"youtube"`
	Links ConfigQuickLinks                `yaml:"quicklinks"`
}

type ConfigProperties struct {
	Port            int    `yaml:"port"`
	RefreshInterval int    `yaml:"refreshInterval"`
	RetryInterval   int    `yaml:"retryInterval"`
	DatabasePath    string `yaml:"dbPath`
}

type ConfigTitleURLItem struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}

type ConfigQuickLinks struct {
	List1Name string               `yaml:"list1Name"`
	List2Name string               `yaml:"list2Name"`
	List3Name string               `yaml:"list3Name"`
	List4Name string               `yaml:"list4Name"`
	List1     []ConfigTitleURLItem `yaml:"list1"`
	List2     []ConfigTitleURLItem `yaml:"list2"`
	List3     []ConfigTitleURLItem `yaml:"list3"`
	List4     []ConfigTitleURLItem `yaml:"list4"`
}
