package netutils

import (
	"com.dk.gateway/src/common/logging"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

func LongPolling(watchUrl string, timeout time.Duration) ([]byte, error) {

	logging.Info("LongPolling", zap.String("watch_url", watchUrl))

	for {
		client := http.Client{Timeout: timeout}
		resp, err := client.Get(watchUrl)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 304 {
			logging.Info("LongPolling not modified")
			_ = resp.Body.Close()
			continue
		}

		if resp.StatusCode != 200 {
			logging.Error("LongPolling status code not 200",
				zap.Int("code", resp.StatusCode))
			_ = resp.Body.Close()
			time.Sleep(time.Second * 1)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return data, err
	}
}
