package reverseproxy

import (
	"fmt"
	"net/url"
	"oauth2-proxy-nexus3/authprovider"
	"oauth2-proxy-nexus3/authprovider/gitlab"
	"oauth2-proxy-nexus3/authprovider/oicd_generic"
)

func newAuthproviderClient(authProvider string, URL *url.URL) (authprovider.Client, error) {
	switch authProvider {
	case "oicd_generic":
		return &oicd_generic.Client{URL: URL}, nil
	case "gitlab":
		return &gitlab.Client{URL: URL}, nil
	default:
		return nil, fmt.Errorf("%s AuthProvider not available", authProvider)
	}
}
