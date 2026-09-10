package config

import (
	"jobqueue/pkg/server"
	"os"
)

var (
	// Data hold all configuration data
	Data config = getConfig()
)

type config struct {
	Server server.Config
}

func getConfig() config {
	env := os.Getenv("MY_ENV")
	ConfigData := config{}
	switch env {
	case "staging":
		ConfigData.Server = server.Config{
			Port: 58577,
		}
	case "production":
		ConfigData.Server = server.Config{
			Port: 58578,
		}
	default:
		ConfigData.Server = server.Config{
			Port: 58579,
		}
	}
	return ConfigData
}
