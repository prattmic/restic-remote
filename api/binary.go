package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

// Binary describes an executable that can be downloaded.
type Binary struct {
	// Name is the descriptive name of the binary.
	Name string

	// Path is the relative path where this binary can be downloaded.
	//
	// The path is relative to the bucket/URL root containing all binaries.
	Path string

	// Version is the version of the binary.
	Version string
}

// GetBinary gets the binary info for name.
func (a *API) GetBinary(name string) (*Binary, error) {
	v := url.Values{}
	v.Set("name", name)
	u := a.url(binaryEndpoint)
	u.RawQuery = v.Encode()

	r, err := a.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error making binary request: %v", err)
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body of response %+v: %v", r, err)
	}

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return nil, fmt.Errorf("error response when getting binary: %+v\n%s", r, string(b))
	}

	var binary Binary
	if err := json.Unmarshal(b, &binary); err != nil {
		return nil, fmt.Errorf("error unmarshalling binary %q: %v", string(b), err)
	}

	return &binary, nil
}

// PostBinary posts a new binary info.
func (a *API) PostBinary(b *Binary) error {
	if err := a.postJSON(a.url(binaryEndpoint), b); err != nil {
		return fmt.Errorf("error writing binary %+v: %v", b, err)
	}
	return nil
}
