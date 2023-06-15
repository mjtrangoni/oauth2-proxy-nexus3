package oicd_generic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
)

// NewTestServer returns an `httptest.Server` that partially implements the OpenID Connect API.
func NewTestServer(accessToken string, userInfo *UserInfo) *httptest.Server {
	userInfoEndpoingPathRegex := regexp.MustCompile(`/auth/realms/([\w\-]+)/protocol/openid-connect/userinfo`)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userInfoEndpoingPathRegex.MatchString(r.URL.Path) {
			if r.Header["Authorization"][0] == "Bearer "+accessToken {
				payload, _ := json.Marshal(&userInfo)

				w.WriteHeader(200)
				w.Write(payload)
			} else {
				w.WriteHeader(401)
			}
		} else {
			w.WriteHeader(404)
		}
	}))
}
