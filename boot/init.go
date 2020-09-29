package boot

import (
	"com.dk.gateway/src/common/logging"
	"com.dk.gateway/src/config"
	"com.dk.gateway/src/config/localcfg"
	"com.dk.gateway/src/route"
	"flag"
	"fmt"
	"os"
	"path"
)

func Boot()  {
	workDir := flag.String("work_dir", "", "--word_dir project root direction")
	cfgFile := flag.String("conf", "", "--conf config file")
	flag.Parse()

	if *workDir == "" || *cfgFile == "" {
		panic("run: ./gateway --work_dir dir_name --conf config_file")
	}

	cfgLoader, err := localcfg.NewCfgLoader(*cfgFile)
	if err != nil {
		panic(err.Error())
	}

	logDir := path.Join(*workDir, "log")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		panic(err.Error())
	}

	logCfg := cfgLoader.GetLogCfg()
	logging.LogInit(path.Join(logDir, logCfg.FileName),
		logCfg.MaxSize, logCfg.MaxBackups, logCfg.MaxAge, logCfg.Level)

	hotCfg := config.NewHotConfig(cfgLoader.GetApolloCfg())
	err = hotCfg.Start()
	if err != nil {
		panic(fmt.Sprintf("init config failed %s", err.Error()))
	}

	route.Route()
}
