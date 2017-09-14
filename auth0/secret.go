// Copyright (c) 2016 Yannick Heinrich
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package auth0

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	gauth0 "github.com/auth0-community/go-auth0"
	jose "gopkg.in/square/go-jose.v2"
)

// SecretProvider implements gauth0.SecretProvider. It is nearly identical to
// gauth0.JWKClient, except that it works on App Engine.
type SecretProvider struct {
	// jwksURL is the URL from which to download keys.
	jwksURL string

	// mu protects the fields below.
	mu sync.Mutex

	// keys contains the known keys, which are downloaded as needed.
	keys map[string]jose.JSONWebKey
}

// NewSecretProvider returns a SecretProvider that fetchs keys from the
// provided JWKS URL.
func NewSecretProvider(url string) *SecretProvider {
	return &SecretProvider{
		jwksURL: url,
		keys:    make(map[string]jose.JSONWebKey),
	}
}

// getKey looks up the key for ID.
func (s *SecretProvider) getKey(ctx context.Context, ID string) (jose.JSONWebKey, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exist := s.keys[ID]

	if !exist {
		// TODO(prattmic): don't continuously re-download if we keep
		// receiving invalid keys.
		if err := s.downloadKeys(ctx); err != nil {
			log.Printf("Error downloading keys: %v", err)
		}
	}

	key, exist = s.keys[ID]
	return key, exist
}

// downloadKeys fetchs keys from the JWKS URL.
func (s *SecretProvider) downloadKeys(ctx context.Context) error {
	c := httpClient(ctx)

	r, err := c.Get(s.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", s.jwksURL, err)
	}
	defer r.Body.Close()

	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return gauth0.ErrInvalidContentType
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var jwks gauth0.JWKS
	if err = json.Unmarshal(b, &jwks); err != nil {
		return fmt.Errorf("failed to unmarshal %q: %v", string(b), err)
	}

	if len(jwks.Keys) < 1 {
		return gauth0.ErrNoKeyFound
	}

	for _, key := range jwks.Keys {
		s.keys[key.KeyID] = key
	}

	return nil
}

// GetSecret implements gauth0.SecretProvider.GetSecret.
func (s *SecretProvider) GetSecret(req *http.Request) (interface{}, error) {
	ctx := requestContext(req)

	t, err := gauth0.FromHeader(req)
	if err != nil {
		return nil, err
	}

	if len(t.Headers) < 1 {
		return nil, gauth0.ErrInvalidTokenHeader
	}

	header := t.Headers[0]
	if header.Algorithm != "RS256" {
		return nil, gauth0.ErrInvalidAlgorithm
	}

	key, ok := s.getKey(ctx, header.KeyID)
	if !ok {
		return nil, gauth0.ErrNoKeyFound
	}

	return key.Key, nil
}
