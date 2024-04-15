package reverseproxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"oauth2-proxy-nexus3/config"
	"oauth2-proxy-nexus3/logger"
	"oauth2-proxy-nexus3/nexus"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ReverseProxy It represents the reverse proxy which
// is the "glue" between oauth2-proxy, the Auth provider and Nexus 3.
type ReverseProxy struct {
	Router *chi.Mux
	Redis  *redis.Client
	Nexus  nexus.Client
}

// writeErrCb write error to the response and the logs
func writeErrCb(w http.ResponseWriter, msg string, code int) {
	logger.Error("HTTP Server", zap.Int("code", code), zap.String("msg", msg))
	http.Error(w, msg, code)
}

// Run initializes and returns a new `ReverseProxy`.
func Run(cfg *config.Config) *ReverseProxy {
	var (
		ctx      = context.Background()
		username = ""
	)

	s := ReverseProxy{
		Router: chi.NewRouter(),
		Redis: redis.NewClient(
			&redis.Options{
				Addr:     cfg.RedisConnectionURL,
				Password: cfg.RedisPassword,
				DB:       10,
			},
		),
		Nexus: nexus.Client{
			BaseURL:  cfg.NexusURL,
			Username: cfg.NexusAdminUser,
			Password: cfg.NexusAdminPassword,
		},
	}

	s.Router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cfg.OAuth2ProxyCookieName)
		if err == nil {
			logger.Debug("request info",
				zap.String("oauth2_proxy cookie value", cookie.Value),
				zap.String("provider", cfg.AuthProvider))
			username, err = s.Redis.Get(ctx, cookie.Value).Result()
			if err == redis.Nil {
				var accessToken = r.Header.Get(cfg.AuthProviderAccessTokenHeader)
				if accessToken == "" {
					writeErrCb(w, "header "+cfg.AuthProviderAccessTokenHeader+
						" value is null", http.StatusBadRequest)

					return
				}

				authproviderClient, err := newAuthproviderClient(cfg.AuthProvider, cfg.AuthProviderURL)
				if err != nil {
					writeErrCb(w, err.Error(), http.StatusNotImplemented)

					return
				}

				userInfo, err := authproviderClient.GetUserInfo(accessToken)
				if err != nil {
					writeErrCb(w, err.Error(), http.StatusInternalServerError)

					return
				}
				logger.Debug("UserInfo from provider client", zap.Any("info", userInfo),
					zap.String("provider", cfg.AuthProvider))

				if err = s.Nexus.SyncUser(
					userInfo.Username(),
					userInfo.Givenname(),
					userInfo.Familyname(),
					userInfo.EmailAddress(),
					userInfo.Roles(),
				); err != nil {
					writeErrCb(w, err.Error(), http.StatusInternalServerError)

					return
				}
				// set cookie in redis
				username := userInfo.Username()
				err = s.Redis.Set(ctx, cookie.Value, username, time.Hour*time.Duration(cfg.RedisTTLHours)).Err()
				if err != nil {
					writeErrCb(w, "couldn't store oauth2_proxy cookie in redis"+err.Error(),
						http.StatusBadRequest)

					return
				}
			}
		} else {
			writeErrCb(w, "couldn't get the oauth2_proxy cookie",
				http.StatusBadRequest)

			return
		}

		r.Header.Set(cfg.NexusRutHeader, username)

		httputil.NewSingleHostReverseProxy(cfg.NexusURL).ServeHTTP(w, r)
	})

	return &s
}
