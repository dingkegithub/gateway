package config

type CfgBackend interface {
	Start()

	Load() (interface{}, error)

	Register(func(interface{}))
}

