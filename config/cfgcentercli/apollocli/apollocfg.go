package apollocli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/dingkegithub/gateway/common/logging"
	"github.com/dingkegithub/gateway/config/cfgcentercli"
	"github.com/dingkegithub/gateway/config/localcfg"
	"github.com/dingkegithub/gateway/utils/netutils"
	"github.com/dingkegithub/gateway/utils/osutils"
	"go.uber.org/zap"
)

type ApolloCfg struct {
	synUrl        string
	watchUrl      string
	cacheUrl      string
	localBak      string
	notifications []*Notification
	namespaces    map[string]int32
	param         *localcfg.ApolloParam
	listeners     map[string]chan<- interface{}
	mutx          sync.RWMutex
	memCfg        map[string]map[string]string
	closeChan     chan struct{}
}

func NewApolloCfg(p *localcfg.ApolloParam) cfgcentercli.CfgCenterCli {

	// {config_server_url}/notifications/v2?appId={appId}&cluster={clusterName}&notifications={notifications}
	watcherUrl := fmt.Sprintf("%s/notifications/v2/?appId=%s&cluster=%s", p.CfgServer, p.AppId, p.Cluster)

	// db url: {config_server_url}/configs/{appId}/{clusterName}/{namespaceName}?releaseKey={releaseKey}&ip={clientIp}
	synCfgUrl := fmt.Sprintf("%s/configs/%s/%s", p.CfgServer, p.AppId, p.Cluster)

	// cache url: {config_server_url}/configfiles/json/{appId}/{clusterName}/{namespaceName}
	cacheUrl := fmt.Sprintf("%s/configfiles/json/%s/%s", p.CfgServer, p.AppId, p.Cluster)

	namespaces := make(map[string]int32)
	notifications := make([]*Notification, 0, len(p.NameSpaces))

	for _, v := range p.NameSpaces {
		notifications = append(notifications,
			&Notification{
				NamespaceName:  v,
				NotificationId: -1,
			})
		namespaces[v] = -1
	}

	return &ApolloCfg{
		param:         p,
		watchUrl:      watcherUrl,
		synUrl:        synCfgUrl,
		cacheUrl:      cacheUrl,
		namespaces:    namespaces,
		notifications: notifications,
		localBak:      p.LocalBak,
		memCfg:        make(map[string]map[string]string),
		closeChan:     make(chan struct{}),
		listeners:     make(map[string]chan<- interface{}),
	}
}

func (ac *ApolloCfg) Open() {
	for {
		err := ac.load()
		if err != nil {
			logging.Error("init load cfg failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		break
	}
	go ac.watch()
}

func (ac *ApolloCfg) Close() {
	ac.closeChan <- struct{}{}
	time.Sleep(time.Millisecond)
	<-ac.closeChan
}

func (ac *ApolloCfg) Register(key string, listener chan<- interface{}) {
	ac.listeners[key] = listener
}

func (ac *ApolloCfg) Get(key, ns string) string {
	ac.mutx.RLock()
	defer ac.mutx.RUnlock()

	nm := ns
	if ns == "" {
		nm = "application"
	}
	if cfg, ok := ac.memCfg[nm]; ok {
		if key == "" {
			return ""
		}

		if kCfg, ok := cfg[key]; ok {
			return kCfg
		}
	}
	return ""
}

func (ac *ApolloCfg) onChange() {
	for _, listener := range ac.listeners {
		listener <- struct{}{}
	}
}

func (ac *ApolloCfg) loadFromLocalCache() ([]byte, error) {
	return osutils.Read(ac.localBak)
}

func (ac *ApolloCfg) flushToLocalCache(content []byte) error {
	bakFile := fmt.Sprintf("%s.swp", ac.localBak)
	err := osutils.Write(bakFile, content)
	if err != nil {
		return err
	}

	err = os.Rename(bakFile, ac.localBak)
	if err != nil {
		return err
	}

	return nil
}

func (ac *ApolloCfg) load() error {
	err := ac.pollingLoad()
	if err != nil {
		content, err := ac.loadFromLocalCache()
		if err != nil {
			logging.Error("load from local cache config exception", zap.Error(err))
			return err
		}

		err = json.Unmarshal(content, ac.memCfg)
		if err != nil {
			logging.Error("unmarshal cache config exception", zap.Error(err))
			return err
		}
		return nil
	} else {
		c, err := json.Marshal(ac.memCfg)
		if err == nil {
			ac.flushToLocalCache(c)
		}
		return nil
	}
}

func (ac *ApolloCfg) urlEncodeNotification() string {
	n := make([]*Notification, 0, len(ac.namespaces))
	for ns, id := range ac.namespaces {
		n = append(n, &Notification{
			NamespaceName:  ns,
			NotificationId: id,
		})
	}

	jsData, err := json.Marshal(n)
	if err != nil {
		logging.Error("config.urlEncodeNotification", zap.Error(err))
		return ""
	}
	logging.Debug("url encode notification", zap.String("notifications", string(jsData)))
	return url.QueryEscape(string(jsData))
}

func (ac *ApolloCfg) pollingLoad() error {
	nxtQueryParams := ac.urlEncodeNotification()
	if nxtQueryParams == "" {
		logging.Error("config.watch params", zap.String("nxtQueryParams", nxtQueryParams))
		return InvalidParamErr
	}

	watchUrl := fmt.Sprintf("%s&notifications=%s", ac.watchUrl, nxtQueryParams)
	body, err := netutils.LongPolling(watchUrl, 65*time.Second)
	if err != nil {
		logging.Error("config.watch polling connect", zap.Error(err))
		return err
	}

	var apolloPollingData []*ApolloPollingData
	err = json.Unmarshal(body, &apolloPollingData)
	if err != nil {
		logging.Error("config.watch unmarshal", zap.Error(err))
		return err
	}

	for _, v := range apolloPollingData {
		ac.namespaces[v.NamespaceName] = v.NotificationId

		synUrl := fmt.Sprintf("%s/%s", ac.synUrl, v.NamespaceName)
		res, err := ac.Syn(synUrl)
		if err == nil {
			ac.mutx.Lock()
			ac.memCfg[v.NamespaceName] = res
			ac.mutx.Unlock()
		}
	}

	return nil
}

func (ac *ApolloCfg) watch() {
	for {
		err := ac.pollingLoad()
		if err == nil {
			ac.onChange()
		}
	}
}

func (ac *ApolloCfg) Syn(synUrl string) (map[string]string, error) {

	logging.Info("apollo.syn", zap.String("url", synUrl))

	resp, err := http.Get(synUrl)
	if err != nil {
		logging.Info("apollo.Syn read", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != 200 {
		logging.Error("apollo.syn error status code",
			zap.Int("code", resp.StatusCode))
		return nil, HttpNot200Err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Error("apollo.Syn http body", zap.Error(err))
		return nil, err
	}
	fmt.Println("apollo.syn read body: ", string(content))

	apolloSynRespData := &ApolloSynRespData{}
	err = json.Unmarshal(content, &apolloSynRespData)
	if err != nil {
		logging.Error("config.Syn parse body", zap.Error(err))
		return nil, err
	}

	return apolloSynRespData.Configurations, nil
}
