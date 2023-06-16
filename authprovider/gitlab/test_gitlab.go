package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"oauth2-proxy-nexus3/logger"

	"go.uber.org/zap"
)

// NewTestServer returns an `httptest.Server` that partially implements the GitLab OIDC API.
func NewTestServer(accessToken string, userInfo *UserInfo) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == userInfoEndpointPath {
			if r.Header["Authorization"][0] == "Bearer "+accessToken {
				payload, _ := json.Marshal(&userInfo)

				w.WriteHeader(200)
				_, err := w.Write(payload)
				if err != nil {
					logger.Error("gitlab provider payload write", zap.Error(err))
				}
			} else {
				w.WriteHeader(401)
			}
		} else {
			w.WriteHeader(404)
		}
	}))
}
