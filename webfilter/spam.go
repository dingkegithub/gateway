package webfilter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dingkegithub/gateway/common"
	"github.com/dingkegithub/gateway/common/logging"
	"github.com/dingkegithub/gateway/config"
	"github.com/dingkegithub/gateway/utils/netutils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/willf/bloom"
	"go.uber.org/zap"
)

type SpamCfg struct {
	IntervalPullBlacklist uint64             `json:"interval_pull_blacklist"`
	Cache                 *common.RedisParam `json:"cache"`
	Bloom                 *common.Bloom      `json:"bloom"`
}

type Antispam struct {
	mutex       sync.RWMutex
	cfg         *SpamCfg
	cfgName     string
	cli         *redis.Client
	cfgListener chan interface{}
	blacklist   *bloom.BloomFilter
	cfgCenter   *config.HotCfg
}

func NewAntispam() *Antispam {

	cfgCenterCli := config.Instance()

	spam := &Antispam{
		cfgName:     "spam",
		cfgCenter:   cfgCenterCli,
		cfgListener: make(chan interface{}),
	}

	spam.cfgUpdate()
	go spam.listenCfgCenter()
	return spam
}

func (spam *Antispam) Register() {
	spam.cfgCenter.Register("", spam.cfgListener)
}

func (spam *Antispam) SpamCheck(c *gin.Context) {
	remoteHost := strings.Split(c.Request.RemoteAddr, ":")[0]

	var uid string

	cookies := c.Request.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == common.REQ_ID {
			uid = cookie.Value
		}
	}

	if spam.isBlack(remoteHost, uid) {
		netutils.StdResponse(c, http.StatusForbidden, netutils.BuildResponseBody(-1, "you had been set black user", nil))
		c.Abort()
		return
	}

	c.Next()
}

func (spam *Antispam) isBlack(names ...string) bool {
	spam.mutex.RLock()
	defer spam.mutex.RUnlock()
	for _, name := range names {
		if name == "" {
			continue
		}
		if spam.blacklist.TestString(name) {
			return true
		}
	}

	return false
}

func (spam *Antispam) listenCfgCenter() {
	blacklistTick := time.Tick(time.Duration(spam.cfg.IntervalPullBlacklist) * time.Second)
	for {
		select {
		case <-spam.cfgListener:
			logging.Info("cfg notify")
			spam.cfgUpdate()

		case <-blacklistTick:
			logging.Info("start update blacklist")
			spam.pollingPullBlacklist()
		}
	}
}

func (spam *Antispam) cfgUpdate() {
	cfgStr := spam.cfgCenter.Get(spam.cfgName, "")

	var spamCfg *SpamCfg
	err := json.Unmarshal([]byte(cfgStr), &spamCfg)
	if err != nil {
		logging.Info("parse new cfg failed", zap.Error(err))
		return
	}

	cli := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", spamCfg.Cache.Host, spamCfg.Cache.Port),
		DB:         spamCfg.Cache.Db,
		MaxRetries: 3,
	})

	res, err := cli.SMembers(context.Background(), common.ANTISPAM_BLACKLIST).Result()
	if err != nil {
		logging.Info("load antispam blacklist failled", zap.Error(err))
		return
	}

	blacklist := bloom.New(spamCfg.Bloom.Cap, spamCfg.Bloom.HashNum)
	for _, v := range res {
		logging.Debug("insert balck item", zap.String("item", v))
		blacklist.AddString(v)
	}

	spam.mutex.Lock()
	spam.cfg = spamCfg
	if spam.cli != nil {
		spam.cli.Close()
	}
	spam.cli = cli
	spam.blacklist = blacklist
	spam.mutex.Unlock()
}

func (spam *Antispam) pollingPullBlacklist() {
	spam.mutex.RLock()
	res, err := spam.cli.SMembers(context.Background(), common.ANTISPAM_BLACKLIST).Result()
	if err != nil {
		logging.Info("load antispam blacklist failled", zap.Error(err))
		spam.mutex.RUnlock()
		return
	}
	spam.mutex.RUnlock()

	blacklist := bloom.New(spam.cfg.Bloom.Cap, spam.cfg.Bloom.HashNum)
	for _, v := range res {
		logging.Debug("insert balck item", zap.String("item", v))
		blacklist.AddString(v)
	}

	spam.mutex.Lock()
	defer spam.mutex.Unlock()
	spam.blacklist = blacklist
}
