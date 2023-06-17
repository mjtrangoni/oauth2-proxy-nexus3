package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"oauth2-proxy-nexus3/config"
	"oauth2-proxy-nexus3/logger"
	"oauth2-proxy-nexus3/nexus"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const routeName = "main"

// ReverseProxy It represents the reverse proxy which
// is the "glue" between oauth2-proxy, the Auth provider and Nexus 3.
type ReverseProxy struct {
	Router *mux.Router
}

// New initializes and returns a new `ReverseProxy`.
func New(cfg *config.Config) *ReverseProxy {
	s := ReverseProxy{
		Router: mux.NewRouter().StrictSlash(true),
	}

	nexusClient := nexus.Client{
		BaseURL:  cfg.NexusURL,
		Username: cfg.NexusAdminUser,
		Password: cfg.NexusAdminPassword,
	}

	s.Router.
		PathPrefix("/").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				writeErrCb = func(msg string, code int) {
					logger.Error("HTTP Server", zap.Int("code", code), zap.String("msg", msg))
					http.Error(w, msg, code)
				}

				accessToken = r.Header.Get(cfg.AuthProviderAccessTokenHeader)
			)
			cookie, err := r.Cookie(cfg.OAuth2ProxyCookieName)
			if err == nil {
				logger.Debug("request info",
					zap.String("oauth2_proxy cookie value", cookie.Value),
					zap.String("provider", cfg.AuthProvider))
			} else {
				logger.Debug("couldn't get oauth2_proxy cookie value")
				logger.Debug("request info",
					zap.Any("header", r.Header),
					zap.String("provider", cfg.AuthProvider))
			}

			if accessToken == "" {
				writeErrCb("header "+cfg.AuthProviderAccessTokenHeader+
					" value is null", http.StatusBadRequest)

				return
			}

			authproviderClient, err := newAuthproviderClient(cfg.AuthProvider, cfg.AuthProviderURL)
			if err != nil {
				writeErrCb(err.Error(), http.StatusNotImplemented)

				return
			}

			userInfo, err := authproviderClient.GetUserInfo(accessToken)
			if err != nil {
				writeErrCb(err.Error(), http.StatusInternalServerError)

				return
			}
			logger.Debug("UserInfo from provider client", zap.Any("info", userInfo),
				zap.String("provider", cfg.AuthProvider))

			if err = nexusClient.SyncUser(
				userInfo.Username(),
				userInfo.EmailAddress(),
				userInfo.Roles(),
			); err != nil {
				writeErrCb(err.Error(), http.StatusInternalServerError)

				return
			}

			r.Header.Set(cfg.NexusRutHeader, userInfo.Username())

			httputil.NewSingleHostReverseProxy(cfg.NexusURL).ServeHTTP(w, r)
		}).
		Name(routeName)

	return &s
}
