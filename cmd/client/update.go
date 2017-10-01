package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

	"cloud.google.com/go/storage"
	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/log"
	"github.com/prattmic/restic-remote/restic"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func updateCheck(a *api.API, r *restic.Restic) error {
	release, err := a.GetRelease()
	if err != nil {
		return fmt.Errorf("error getting current release: %v", err)
	}

	log.Infof("Latest release: %+v", release)

	rver, err := r.Version()
	if err != nil {
		return fmt.Errorf("error getting restic version: %v", err)
	}

	log.Infof("Current restic version: %s", rver)
	log.Infof("Current client version: %s", versionStr)

	updateRestic := rver != release.ResticVersion
	updateClient := versionStr != release.ClientVersion
	if !updateRestic && !updateClient {
		log.Infof("No updates available")
		return nil
	}

	return performUpdate(release, updateRestic, updateClient)
}

func performUpdate(release *api.Release, updateRestic, updateClient bool) error {
	dir, err := ioutil.TempDir("", "restic-remote-update")
	if err != nil {
		return fmt.Errorf("error creating update directory: %v", err)
	}
	defer os.RemoveAll(dir)

	bucketURL := viper.GetString("google.binary-bucket")
	if bucketURL == "" {
		return fmt.Errorf("Binary bucket not configured")
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		return fmt.Errorf("malformed bucket %s", bucketURL)
	}
	bucket := u.Host

	resticBin := "restic"
	clientBin := "client"
	if runtime.GOOS == "windows" {
		resticBin += ".exe"
		clientBin += ".exe"
	}

	tmpRestic := filepath.Join(dir, resticBin)
	tmpClient := filepath.Join(dir, clientBin)

	srcRestic := path.Join(release.Path, resticBin)
	srcClient := path.Join(release.Path, clientBin)

	ctx := context.Background()
	creds := viper.GetString("restic.google-credentials")  // TODO: not restic.
	c, err := storage.NewClient(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		return fmt.Errorf("error creating storage client: %v", err)
	}

	bkt := c.Bucket(bucket)

	if updateRestic {
		if err := download(ctx, tmpRestic, bkt, srcRestic); err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}
	if updateClient {
		if err := download(ctx, tmpClient, bkt, srcClient); err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}

	return nil
}

func download(ctx context.Context, dst string, bkt *storage.BucketHandle, path string) error {
	log.Infof("Downloading %s to %s", path, dst)

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("error opening destination: %v", err)
	}
	defer f.Close()

	obj := bkt.Object(path)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("error opening object %s: %v", path, err)
	}
	defer r.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("error downloading object: %v", err)
	}

	return nil
}
