package main

import (
	"context"
	"fmt"
	"path"

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

	glog.Infof("Rolling out restic...")
	bin := api.Binary{
		Name:    "restic",
		Path:    path.Join(ver.release, "restic"),
		Version: ver.restic,
	}
	if err := a.PostBinary(&bin); err != nil {
		return fmt.Errorf("error POSTing restic: %v", err)
	}

	glog.Infof("Rolling out client...")
	bin = api.Binary{
		Name:    "client",
		Path:    path.Join(ver.release, "client"),
		Version: ver.client,
	}
	if err := a.PostBinary(&bin); err != nil {
		return fmt.Errorf("error POSTing client: %v", err)
	}

	return nil
}
