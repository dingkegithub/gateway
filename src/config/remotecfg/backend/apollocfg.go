package backend

import (
	"com.dk.gateway/src/common/logging"
	"com.dk.gateway/src/config"
	"com.dk.gateway/src/config/localcfg"
	"com.dk.gateway/src/utils/netutils"
	"encoding/json"
	"fmt"
	"github.com/fatih/set"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type ApolloCfg struct {
	param         *localcfg.ApolloParam
	listeners     *set.Set
	watchUrl      string
	synUrl        string
	notifications []*Notification
}

func NewApolloCfg(p *localcfg.ApolloParam) config.CfgBackend {

	watcherUrl := fmt.Sprintf("%s/notifications/v2/?appId=%s&cluster=%s",
		p.CfgServer, p.AppId, p.Cluster)

	synCfgUrl := fmt.Sprintf("%s/configs/%s/%s",
		p.CfgServer, p.AppId, p.Cluster)

	var notifications []*Notification
	for _, v := range p.NameSpaces {
		notifications = append(notifications,
			&Notification{
				NamespaceName:  v,
				NotificationId: -1,
			})
	}

	return &ApolloCfg{
		param:         p,
		watchUrl:      watcherUrl,
		synUrl:        synCfgUrl,
		notifications: notifications,
		listeners: set.New(set.ThreadSafe).(*set.Set),
	}
}

func (ac *ApolloCfg) Start() {
	ac.watch()
}

func (ac *ApolloCfg) Register(listener func(interface{}))  {
	ac.listeners.Add(listener)
}

func (ac *ApolloCfg) onChange(data interface{})  {
	ac.listeners.Each(func(item interface{}) bool {
		switch item.(type) {
		case func([]byte):
			listener := item.(func(interface{}))
			listener(data)
		default:
			logging.Error("apollo cfg.onChange unknown listener")
		}
		return true
	})
}

func (ac *ApolloCfg) Load() (interface{}, error) {

	nameCfg := make(map[string]map[string]string)

	for _, v := range ac.notifications {
		synUrl := fmt.Sprintf("%s/%s", ac.synUrl, v.NamespaceName)
		res, err := ac.Syn(synUrl)
		if err == nil {
			nameCfg[v.NamespaceName] = res
		}
	}

	return nameCfg, nil
}


func (ac *ApolloCfg) urlEncodeNotification(n []*Notification) string {
	jsData, err := json.Marshal(n)
	if err != nil {
		logging.Error("config.urlEncodeNotification", zap.Error(err))
		return ""
	}
	return url.QueryEscape(string(jsData))
}

func (ac *ApolloCfg) watch() {

	go func() {
		nxtQueryParams := ac.urlEncodeNotification(ac.notifications)

		for {
			if nxtQueryParams == "" {
				logging.Error("config.watch params", zap.String("nxtQueryParams", nxtQueryParams))
				return
			}

			watchUrl := fmt.Sprintf("%s&notifications=%s", ac.watchUrl, nxtQueryParams)
			fmt.Println("watch url: ", watchUrl)
			body, err := netutils.LongPolling(ac.watchUrl, 65 * time.Second)
			if err != nil {
				logging.Error("config.watch polling connect", zap.Error(err))
				continue
			}

			var apolloPollingData []*ApolloPollingData
			err = json.Unmarshal(body, &apolloPollingData)
			if err != nil {
				logging.Error("config.watch unmarshal", zap.Error(err))
				continue
			}

			cfgMap := make(map[string]map[string]string)
			notifications := make([]*Notification, 0, len(apolloPollingData))
			for _, v := range apolloPollingData {
				notification := &Notification{
					NamespaceName:  v.NamespaceName,
					NotificationId: v.NotificationId,
				}
				notifications = append(notifications, notification)

				synUrl := fmt.Sprintf("%s/%s", ac.synUrl, v.NamespaceName)
				res, err := ac.Syn(synUrl)
				if err == nil {
					cfgMap[v.NamespaceName] = res
				}
			}

			nxtQueryParams = ac.urlEncodeNotification(notifications)
			ac.onChange(cfgMap)
		}
	}()
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
