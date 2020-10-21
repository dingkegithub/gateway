package boot

import (
	"flag"
	"os"
	"path"

	"github.com/dingkegithub/gateway/common/logging"
	"github.com/dingkegithub/gateway/config"
	"github.com/dingkegithub/gateway/config/localcfg"
	"github.com/dingkegithub/gateway/route"
)

func Boot() {
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

	logging.Info("initing config center")
	config.NewHotCfg(cfgLoader.GetApolloCfg())
	logging.Info("inited config center")

	route.Route()
}
