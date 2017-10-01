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
	"github.com/prattmic/restic-remote/binver"
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

func download(ctx context.Context, dst *os.File, bkt *storage.BucketHandle, path string) error {
	log.Infof("Downloading %s to %s", path, dst.Name())

	obj := bkt.Object(path)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("error opening object %s: %v", path, err)
	}
	defer r.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return fmt.Errorf("error downloading object: %v", err)
	}

	return nil
}

func tempExecutable(dir, prefix string) (*os.File, error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return f, fmt.Errorf("error creating tmpfile: %v", err)
	}

	if err := f.Chmod(0755); err != nil {
		f.Close()
		return nil, fmt.Errorf("error setting permissions: %v", err)
	}

	return f, nil
}

func performUpdate(release *api.Release, updateRestic, updateClient bool) error {
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

	ctx := context.Background()
	creds := viper.GetString("restic.google-credentials")  // TODO: not restic.
	c, err := storage.NewClient(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		return fmt.Errorf("error creating storage client: %v", err)
	}

	bkt := c.Bucket(bucket)

	var tmpRestic, tmpClient string
	if updateRestic {
		dstRestic := viper.GetString("restic.binary")

		// Create new file in the same folder so we won't accidentally
		// try to do a cross-mount rename later.
		f, err := tempExecutable(filepath.Dir(dstRestic), "restic.new")
		if err != nil {
			return fmt.Errorf("error creating restic tmpfile: %v", err)
		}
		tmpRestic = f.Name()
		defer os.Remove(tmpRestic)

		srcRestic := path.Join(release.Path, resticBin)
		err = download(ctx, f, bkt, srcRestic)
		f.Close()
		if err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}
	if updateClient {
		dstClient, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error finding running executable: %v", err)
		}

		f, err := tempExecutable(filepath.Dir(dstClient), "client.new")
		if err != nil {
			return fmt.Errorf("error creating client tmpfile: %v", err)
		}
		tmpClient = f.Name()
		defer os.Remove(tmpClient)

		srcClient := path.Join(release.Path, clientBin)
		err = download(ctx, f, bkt, srcClient)
		f.Close()
		if err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}

	return checkAndInstall(release, tmpRestic, tmpClient)
}

func checkAndInstall(release *api.Release, tmpRestic, tmpClient string) (err error) {
	// Make sure we got working binaries with the correct versions.

	var dstRestic string
	if tmpRestic != "" {
		version, err := binver.Restic(tmpRestic)
		if err != nil {
			return fmt.Errorf("error getting version of new restic %s: %v", tmpRestic, err)
		}
		if version != release.ResticVersion {
			return fmt.Errorf("restic version mismatch got %q want %q", version, release.ResticVersion)
		}

		// TODO: pass this in?
		dstRestic = viper.GetString("restic.binary")
	}
	if tmpClient != "" {
		version, err := binver.Client(tmpClient)
		if err != nil {
			return fmt.Errorf("error getting version of new client %s: %v", tmpClient, err)
		}
		if version != release.ClientVersion {
			return fmt.Errorf("client version mismatch got %q want %q", version, release.ClientVersion)
		}
	}

	// We'll always need this for execve.
	dstClient, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error finding running executable: %v", err)
	}

	// Move the existing binaries out of the way (and keep them as
	// backups).
	if tmpRestic != "" {
		if err := os.Rename(dstRestic, dstRestic+".old"); err != nil {
			return fmt.Errorf("error moving old restic: %v", err)
		}
		defer func() {
			if err == nil {
				return
			}

			// Something went wrong, revert.
			log.Errorf("Encountered error: %v; moving old restic back", err)
			if err := os.Rename(dstRestic+".old", dstRestic); err != nil {
				log.Errorf("Failed to move restic back: %v", err)
			}
		}()
	}
	if tmpClient != "" {
		if err := os.Rename(dstClient, dstClient+".old"); err != nil {
			return fmt.Errorf("error moving old client: %v", err)
		}
		defer func() {
			if err == nil {
				return
			}

			// Something went wrong, revert.
			log.Errorf("Encountered error: %v; moving old client back", err)
			if err := os.Rename(dstClient+".old", dstClient); err != nil {
				log.Errorf("Failed to move client back: %v", err)
			}
		}()
	}

	// Finally, move the existing binaries into place.
	if tmpRestic != "" {
		if err := os.Rename(tmpRestic, dstRestic); err != nil {
			return fmt.Errorf("error moving new restic: %v", err)
		}
	}
	if tmpClient != "" {
		if err := os.Rename(tmpClient, dstClient); err != nil {
			return fmt.Errorf("error moving new client: %v", err)
		}
	}

	// Success! Re-exec to the new version. This isn't actually needed if
	// we only updated restic, but it is simple enough to restart.
	log.Infof("Updated, restarting...")
	execve(dstClient, os.Args[1:])

	return nil
}
