package oicd_generic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"oauth2-proxy-nexus3/authprovider"
	"oauth2-proxy-nexus3/logger"

	"go.uber.org/zap"
)

// Client implements `authprovider.Client`.
type Client struct {
	URL *url.URL
}

// GetUserInfo implements `authprovider.Client`.
func (s *Client) GetUserInfo(accessToken string) (authprovider.UserInfo, error) {
	endpoint, err := url.Parse(s.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse the OpenID Connect URL: %s", err)
	}

	req, err := http.NewRequest("GET", endpoint.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create the OpenID Connect GET userinfo request: %s", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request the OpenID Connect GET userinfo endpoint on %s: %s", s.URL, err)
	}
	defer res.Body.Close()

	logger.Debug("oidc_generic provider response", zap.Any("response", res.Body), zap.Any("header", res.Header))
	if res.StatusCode != http.StatusOK {
		if resBody, err := io.ReadAll(res.Body); err == nil {
			return nil, fmt.Errorf("failed to request the OpenID Connect GET userinfo endpoint on %s: %s", s.URL, resBody)
		}

		return nil, fmt.Errorf("failed to read the OpenID Connect GET userinfo error response: %s", err)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode the OpenID Connect GET userinfo responses: %s", err)
	}

	return &userInfo, nil
}

// UserInfo implements `authprovider.UserInfo`.
type UserInfo struct {
	User       string   `json:"nickname"`
	GivenName  string   `json:"given_name,omitempty"`
	FamilyName string   `json:"family_name,omitempty"`
	Email      string   `json:"email"`
	Groups     []string `json:"groups"`
}

// Username implements `authprovider.UserInfo`.
func (s *UserInfo) Username() string {
	return s.User
}

// Givenname implements `authprovider.UserInfo`.
func (s *UserInfo) Givenname() string {
	return s.GivenName
}

// Familyname implements `authprovider.UserInfo`.
func (s *UserInfo) Familyname() string {
	return s.FamilyName
}

// EmailAddress implements `authprovider.UserInfo`.
func (s *UserInfo) EmailAddress() string {
	return s.Email
}

// Roles implements `authprovider.UserInfo`.
func (s *UserInfo) Roles() []string {
	return s.Groups
}
