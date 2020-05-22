package config

import (
	"com.dk.gateway/src/config/localcfg"
	"com.dk.gateway/src/config/remotecfg/backend"
	"sync"
)

type HotConfig struct {
	backend CfgBackend
	lock *sync.Mutex
}

func NewHotConfig(p *localcfg.ApolloParam) *HotConfig  {

	bak := backend.NewApolloCfg(p)

	return & HotConfig{
		backend:bak,
		lock: &sync.Mutex{},
	}
}

func (hc *HotConfig) Start() error {
	hc.backend.Register(hc.update)
	data, err := hc.backend.Load()
	if err != nil {
		data, err = hc.loadLocalCfg()
		if err != nil {
			return err
		}
	}
	hc.update(data)
	return nil
}

func (hc *HotConfig) loadLocalCfg() (interface{}, error) {
	return "", nil
}

func (hc *HotConfig) update(data interface{}) {
	// ToDo parse data and update config

	hc.lock.Lock()
	defer hc.lock.Unlock()
}
