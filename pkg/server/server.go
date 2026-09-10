package server

import (
	"fmt"
	"os"

	echo "github.com/labstack/echo/v4"
)

var (
	config Config
)

// Config hold configuration data
type Config struct {
	Port int
}

// Echo hold echo wrapper
type Echo struct {
	Echo *echo.Echo
}

// Start hold echo start wrapper
func (e Echo) Start() error {
	var host string
	env := os.Getenv("MY_ENV")
	if env != "staging" && env != "production" {
		host = "localhost"
	}
	return e.Echo.Start(fmt.Sprintf("%s:%d", host, config.Port))
}

// New generate new echo server
func New(d Config) *Echo {
	// Store the config
	config = d

	return &Echo{echo.New()}
}
