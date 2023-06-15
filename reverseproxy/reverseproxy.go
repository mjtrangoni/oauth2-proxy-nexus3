package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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
func New(
	upstreamURL, authproviderURL, nexusURL *url.URL,
	authprovider, accessTokenHeader, nexusAdminUser, nexusAdminPassword, nexusRutHeader string,
) *ReverseProxy {
	s := ReverseProxy{
		Router: mux.NewRouter().StrictSlash(true),
	}

	nexusClient := nexus.Client{
		BaseURL:  nexusURL,
		Username: nexusAdminUser,
		Password: nexusAdminPassword,
	}

	s.Router.
		PathPrefix("/").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				writeErrCb = func(msg string, code int) {
					logger.Error("HTTP Server", zap.Int("code", code), zap.String("msg", msg))
					http.Error(w, msg, code)
				}

				accessToken = r.Header.Get(accessTokenHeader)
			)

			if accessToken == "" {
				writeErrCb("header "+accessTokenHeader+" value is null", http.StatusBadRequest)

				return
			}

			authproviderClient, err := newAuthproviderClient(authprovider, authproviderURL)
			if err != nil {
				writeErrCb(err.Error(), http.StatusNotImplemented)

				return
			}

			userInfo, err := authproviderClient.GetUserInfo(accessToken)
			if err != nil {
				writeErrCb(err.Error(), http.StatusInternalServerError)

				return
			}

			if err = nexusClient.SyncUser(
				userInfo.Username(),
				userInfo.EmailAddress(),
				userInfo.Roles(),
			); err != nil {
				writeErrCb(err.Error(), http.StatusInternalServerError)

				return
			}

			r.Header.Set(nexusRutHeader, userInfo.Username())

			httputil.NewSingleHostReverseProxy(upstreamURL).ServeHTTP(w, r)
			logger.Debug("HTTP Reverse Proxy call",
				zap.String("username", userInfo.Username()),
				zap.String("email", userInfo.EmailAddress()),
				zap.Strings("roles", userInfo.Roles()),
			)
		}).
		Name(routeName)

	return &s
}
