package cfgcentercli

type CfgCenterCli interface {
	Open()

	Close()

	Register(string, chan<- interface{})

	Get(key, ns string) string
}
