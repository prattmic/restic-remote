package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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

	ctx := context.Background()
	creds := viper.GetString("restic.google-credentials")  // TODO: not restic.
	c, err := storage.NewClient(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		return fmt.Errorf("error creating storage client: %v", err)
	}

	bkt := c.Bucket(bucket)

	var tmpRestic, tmpClient string
	if updateRestic {
		tmpRestic = filepath.Join(dir, resticBin)
		srcRestic := path.Join(release.Path, resticBin)
		if err := download(ctx, tmpRestic, bkt, srcRestic); err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}
	if updateClient {
		tmpClient = filepath.Join(dir, clientBin)
		srcClient := path.Join(release.Path, clientBin)
		if err := download(ctx, tmpClient, bkt, srcClient); err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
	}

	return checkAndInstall(release, tmpRestic, tmpClient)
}

// TODO: dedup
func clientVersion(bin string) (string, error) {
	cmd := exec.Command(bin, "--version")
	b, err := cmd.Output()
	return string(b), err
}

func resticVersion(bin string) (string, error) {
	cmd := exec.Command(bin, "version")
	b, err := cmd.Output()
	return string(b), err
}

func checkAndInstall(release *api.Release, tmpRestic, tmpClient string) (err error) {
	// Make sure we got working binaries with the correct versions.

	var dstRestic, dstClient string
	if tmpRestic != "" {
		version, err := resticVersion(tmpRestic)
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
		version, err := clientVersion(tmpClient)
		if err != nil {
			return fmt.Errorf("error getting version of new client %s: %v", tmpClient, err)
		}
		if version != release.ClientVersion {
			return fmt.Errorf("client version mismatch got %q want %q", version, release.ClientVersion)
		}

		dstClient, err = os.Executable()
		if err != nil {
			return fmt.Errorf("error finding running executable: %v", err)
		}
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
