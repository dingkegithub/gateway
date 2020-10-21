package apollocli

type ApolloSynRespData struct {
	AppId          string            `json:"appId"`
	Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}

type Messages struct {
	Details map[string]interface{} `json:"details"`
}

type ApolloPollingData struct {
	NamespaceName  string   `json:"namespaceName"`
	NotificationId int32    `json:"notificationId"`
	Messages       Messages `json:"messages"`
}

type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int32  `json:"notificationId"`
}

type ApolloPollingParam struct {
	AppId         string          `json:"app_id"`
	Cluster       string          `json:"cluster"`
	Notifications []*Notification `json:"notifications"`
}
