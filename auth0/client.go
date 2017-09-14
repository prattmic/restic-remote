// Package auth0 provides helpers to send Auth0 credentials from a client and
// validate Auth0 credentials on a server.
package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// tokenRequest is the structure used by /oauth/token to request an access
// token using client credentials.
type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
}

// tokenResponse is the structure returned by /oauth/token on success.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint64 `json:"expires_in"`
}

// ClientConfig describes an Auth0 client credentials flow.
type ClientConfig struct {
	// ClientID is the client's id.
	ClientID string `mapstructure:"client-id"`

	// ClientSecret is the client's secret.
	ClientSecret string `mapstructure:"client-secret"`

	// Audience is the unique ID of the target API to access.
	Audience string

	// TokenURL is the /oauth/token URL for the Auth0 account.
	TokenURL string `mapstructure:"token-url"`
}

// tokenSource implements oauth2.TokenSource.
type tokenSource struct {
	ctx  context.Context
	conf ClientConfig
}

// Token implements oauth2.TokenSource.Token.
func (t *tokenSource) Token() (*oauth2.Token, error) {
	treq := tokenRequest{
		GrantType:    "client_credentials",
		ClientID:     t.conf.ClientID,
		ClientSecret: t.conf.ClientSecret,
		Audience:     t.conf.Audience,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&treq); err != nil {
		return nil, fmt.Errorf("failed to get token: error encoding %+v: %v", treq, err)
	}

	req, err := http.NewRequest("POST", t.conf.TokenURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: error creating HTTP: %v", err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: error sending HTTP request %+v: %v", req, err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: error reading response body: %v", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get token: request failed with code %d: %s", res.StatusCode, string(b))
	}

	var tres tokenResponse
	if err := json.Unmarshal(b, &tres); err != nil {
		return nil, fmt.Errorf("failed to get token: error decoding token: %v", err)
	}

	return &oauth2.Token{
		AccessToken: tres.AccessToken,
		TokenType:   tres.TokenType,
		Expiry:      time.Now().Add(time.Duration(tres.ExpiresIn) * time.Second),
	}, nil
}

// NewClient returns an http.Client using a token obtained from Auth0. The
// token will auto-refresh as necessary.
func NewClient(ctx context.Context, conf ClientConfig) (*http.Client, error) {
	if conf.ClientID == "" {
		return nil, fmt.Errorf("client ID must be provided")
	}
	if conf.ClientSecret == "" {
		return nil, fmt.Errorf("client secret must be provided")
	}
	if conf.Audience == "" {
		return nil, fmt.Errorf("audience must be provided")
	}
	if conf.TokenURL == "" {
		return nil, fmt.Errorf("token URL must be provided")
	}

	t := &tokenSource{
		ctx:  ctx,
		conf: conf,
	}
	return oauth2.NewClient(ctx, oauth2.ReuseTokenSource(nil, t)), nil
}
