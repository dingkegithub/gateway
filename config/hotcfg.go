package config

import (
	"sync"

	"github.com/dingkegithub/gateway/config/cfgcentercli"
	"github.com/dingkegithub/gateway/config/cfgcentercli/apollocli"
	"github.com/dingkegithub/gateway/config/localcfg"
)

var hotCfg *HotCfg
var once sync.Once

type HotCfg struct {
	cli cfgcentercli.CfgCenterCli
}

func NewHotCfg(cfgParam *localcfg.ApolloParam) *HotCfg {

	once.Do(func() {
		cli := apollocli.NewApolloCfg(cfgParam)
		cli.Open()
		hotCfg = &HotCfg{
			cli: cli,
		}
	})

	return hotCfg
}

func (hc *HotCfg) Register(ns string, l chan<- interface{}) {
	hc.cli.Register(ns, l)
}

func (hc *HotCfg) Get(key string, ns string) string {
	return hc.cli.Get(key, ns)
}

func Instance() *HotCfg {
	return hotCfg
}
