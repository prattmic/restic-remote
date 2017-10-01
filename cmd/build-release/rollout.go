package main

import (
	"context"
	"fmt"

	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/auth0"
	"github.com/golang/glog"
)

func rolloutRelease(release string, ver *versions) error {
	ctx := context.Background()
	a, err := api.New(ctx, api.Config{
		ClientConfig: auth0.ClientConfig{
			ClientID:     *apiClientID,
			ClientSecret: *apiClientSecret,
			Audience:     *apiAudience,
			TokenURL:     *apiTokenURL,
		},
		Root:         *apiRoot,
	})
	if err != nil {
		return fmt.Errorf("error creating API: %v", err)
	}

	glog.Infof("Rolling out release...")
	rel := api.Release{
		Path:          ver.release,
		ResticVersion: ver.restic,
		ClientVersion: ver.client,
	}
	if err := a.PostRelease(&rel); err != nil {
		return fmt.Errorf("error POSTing release: %v", err)
	}

	return nil
}
