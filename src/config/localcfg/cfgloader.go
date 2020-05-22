package localcfg

import (
	"com.dk.gateway/src/utils/osutils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)


type CfgLoader struct {
	cfgFile string
	localCfg *LocalCfg
}

func NewCfgLoader(f string) (*CfgLoader, error) {

	if ! (osutils.Exists(f) && osutils.IsFile(f)) {
		return nil, errors.New(fmt.Sprintf("file not exist: %s", f))
	}

	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	localCfg := &LocalCfg{}
	err = json.Unmarshal(data, localCfg)
	if err != nil {
		return nil, err
	}

	return &CfgLoader {
		cfgFile:  f,
		localCfg: localCfg,
	}, nil
}

func (cl CfgLoader) GetLogCfg() *Log {
	return cl.localCfg.Log
}

func (cl CfgLoader) GetApolloCfg() *ApolloParam {
	return cl.localCfg.Apollo
}
