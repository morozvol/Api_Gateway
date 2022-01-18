package apiserver

import (
	"net/http"
)

// Start ...
func Start(config *Config) error {
	srv := newServer(config)
	return http.ListenAndServe(config.BindAddr, srv)
}
