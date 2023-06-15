package main

import (
	"crypto/tls"
	"net/http"
	"oauth2-proxy-nexus3/logger"
	"oauth2-proxy-nexus3/reverseproxy"

	env "github.com/caarlos0/env/v8"
	"go.uber.org/zap"
)

var cfg = config{}

func main() {
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Failed to parse the configuration", zap.String("error", err.Error()))
	}

	var (
		reverseProxy = reverseproxy.New(
			cfg.NexusURL, cfg.AuthProviderURL, cfg.NexusURL,
			cfg.AuthProvider, cfg.AuthProviderAccessTokenHeader,
			cfg.NexusAdminUser, cfg.NexusAdminPassword, cfg.NexusRutHeader,
		)

		server = http.Server{Addr: cfg.ListenOn, Handler: reverseProxy.Router}
	)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: cfg.SSLInsecureSkipVerify}

	logger.Info("Starting the proxy")

	if err := server.ListenAndServe(); err != nil {
		logger.Panic(err.Error())
	}
}
