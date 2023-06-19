package reverseproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"oauth2-proxy-nexus3/authprovider/gitlab"
	"oauth2-proxy-nexus3/config"
	"oauth2-proxy-nexus3/nexus"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var (
		oauthAccessToken = "token"
		nexusUser        = nexus.User{
			UserID:       "foo",
			FirstName:    "",
			LastName:     "",
			EmailAddress: "foo@test.bar",
			RoleIDs:      []string{"bar"},
		}
		nexusAvailablesRoles = []nexus.Role{{ID: nexusUser.RoleIDs[0]}}

		gitlabOIDCTestSrv = gitlab.NewTestServer(oauthAccessToken, &gitlab.UserInfo{
			User:       nexusUser.UserID,
			GivenName:  nexusUser.FirstName,
			FamilyName: nexusUser.LastName,
			Email:      nexusUser.EmailAddress,
			Groups:     nexusUser.RoleIDs,
		})
		gitlabOIDCTestSrvURL, _ = url.Parse(gitlabOIDCTestSrv.URL)

		nexusTestSrv = nexus.NewTestServer(
			[]nexus.UserModifier{{User: nexusUser}},
			&nexusAvailablesRoles,
		)
		nexusTestSrvURL, _ = url.Parse(nexusTestSrv.URL)

		rproxyAccessTokenHeader = "X-Forwarded-Access-Token"

		cfg = config.Config{
			NexusURL:                      nexusTestSrvURL,
			AuthProvider:                  "gitlab",
			AuthProviderURL:               gitlabOIDCTestSrvURL,
			AuthProviderAccessTokenHeader: rproxyAccessTokenHeader,
			NexusAdminUser:                "null",
			NexusAdminPassword:            "null",
			NexusRutHeader:                "X-Forwarded-User",
		}
		rproxy = New(&cfg)

		rProxySrv = httptest.NewServer(rproxy.Router.GetRoute(routeName).GetHandler())
	)

	defer gitlabOIDCTestSrv.Close()
	defer nexusTestSrv.Close()
	defer rProxySrv.Close()

	res, err := rProxySrv.Client().Get(rProxySrv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	sucessfulReq, _ := http.NewRequest("GET", rProxySrv.URL, nil)
	sucessfulReq.Header.Add(rproxyAccessTokenHeader, oauthAccessToken)

	res, err = rProxySrv.Client().Do(sucessfulReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}
