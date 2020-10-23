package common

type RedisParam struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
	Db   int    `json:"db"`
}

type Bloom struct {
	HashNum uint `json:"hash_num"`
	Cap     uint `json:"cap"`
}
