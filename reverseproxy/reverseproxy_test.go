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

		cfg = config.Config{
			NexusURL:                      nexusTestSrvURL,
			AuthProvider:                  "gitlab",
			AuthProviderURL:               gitlabOIDCTestSrvURL,
			AuthProviderAccessTokenHeader: "X-Forwarded-Access-Token",
			NexusAdminUser:                "null",
			NexusAdminPassword:            "null",
			NexusRutHeader:                "X-Forwarded-User",
		}
		rproxy = Run(&cfg)

		rProxySrv = httptest.NewServer(rproxy.Router)
	)

	defer gitlabOIDCTestSrv.Close()
	defer nexusTestSrv.Close()
	defer rProxySrv.Close()

	res, err := rProxySrv.Client().Get(rProxySrv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	sucessfulReq, _ := http.NewRequest("GET", rProxySrv.URL, http.NoBody)
	sucessfulReq.Header.Add("X-Forwarded-Access-Token", oauthAccessToken)

	res, err = rProxySrv.Client().Do(sucessfulReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
