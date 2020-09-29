package localcfg

type Log struct {
	Level      int    `json:"level"`
	MaxAge     int    `json:"max_age"`
	MaxSize    int    `json:"max_size"`
	FileName   string `json:"file_name"`
	MaxBackups int    `json:"max_backups"`
}

type ApolloParam struct {
	AppId      string   `json:"app_id"`
	CfgServer  string   `json:"cfg_server"`
	Cluster    string   `json:"cluster"`
	LocalBak   string   `json:"local_bak"`
	NameSpaces []string `json:"name_spaces"`
}

type LocalCfg struct {
	Log    *Log         `json:"log"`
	Apollo *ApolloParam `json:"apollo"`
}
