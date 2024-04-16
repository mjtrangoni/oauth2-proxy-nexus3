package main

import (
	"crypto/tls"
	"net/http"
	"oauth2-proxy-nexus3/config"
	"oauth2-proxy-nexus3/logger"
	"oauth2-proxy-nexus3/reverseproxy"
	"time"

	"runtime/debug"

	env "github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

func main() {
	var Cfg = config.Config{}

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		settings := make(map[string]string)
		for _, kv := range buildInfo.Settings {
			settings[kv.Key] = kv.Value
		}
		logger.Info("Build Info",
			zap.String("go", buildInfo.GoVersion),
			zap.String("path", buildInfo.Path),
			zap.Any("settings", settings),
		)
	} else {
		logger.Error("Couldn't read Build Info")
	}

	if err := env.Parse(&Cfg); err != nil {
		logger.Fatal("Failed to parse the configuration", zap.String("error", err.Error()))
	}

	var (
		reverseProxy = reverseproxy.Run(&Cfg)

		server = http.Server{
			Addr:              Cfg.ListenOn,
			Handler:           reverseProxy.Router,
			ReadHeaderTimeout: 60 * time.Second, // timeout for HTTP requests.
		}
	)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: Cfg.SSLInsecureSkipVerify}

	logger.Info("Starting the proxy")

	if err := server.ListenAndServe(); err != nil {
		logger.Panic(err.Error())
	}
}
