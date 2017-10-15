package main

import (
	"context"
	"fmt"

	"github.com/prattmic/restic-remote/api"
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

func rolloutRelease(release string, ver *versions) error {
	ctx := context.Background()

	var aconf api.Config
	if err := viper.UnmarshalKey("api", &aconf); err != nil {
		return fmt.Errorf("error unmarshalling API config: %v", err)
	}

	a, err := api.New(ctx, aconf)
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
