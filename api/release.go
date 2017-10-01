package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Release describes combined restic/client executable release that can be
// downloaded.
type Release struct {
	// Path is the relative directory path where binaries in this release
	// can be downloaded.
	//
	// The path is relative to the bucket/URL root containing all releases.
	Path string

	// ResticVersion is the version of the restic binary.
	ResticVersion string

	// ClientVersion is the version of the client binary.
	ClientVersion string
}

// GetRelease gets the current release.
func (a *API) GetRelease() (*Release, error) {
	u := a.url(releaseEndpoint)
	r, err := a.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error making release request: %v", err)
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body of response %+v: %v", r, err)
	}

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return nil, fmt.Errorf("error response when getting release: %+v\n%s", r, string(b))
	}

	var release Release
	if err := json.Unmarshal(b, &release); err != nil {
		return nil, fmt.Errorf("error unmarshalling release %q: %v", string(b), err)
	}

	return &release, nil
}

// PostRelease posts new release.
func (a *API) PostRelease(b *Release) error {
	if err := a.postJSON(a.url(releaseEndpoint), b); err != nil {
		return fmt.Errorf("error writing release %+v: %v", b, err)
	}
	return nil
}
