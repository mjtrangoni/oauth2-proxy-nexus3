package reverseproxy

import (
	"fmt"
	"net/url"
	"oauth2-proxy-nexus3/authprovider"
	"oauth2-proxy-nexus3/authprovider/gitlab"
	"oauth2-proxy-nexus3/authprovider/oicd_generic"
)

func newAuthproviderClient(authProvider string, providerURL *url.URL) (authprovider.Client, error) {
	switch authProvider {
	case "oicd_generic":
		return &oicd_generic.Client{URL: providerURL}, nil
	case "gitlab":
		return &gitlab.Client{URL: providerURL}, nil
	default:
		return nil, fmt.Errorf("%s AuthProvider not available", authProvider)
	}
}
