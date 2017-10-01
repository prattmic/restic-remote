package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"cloud.google.com/go/storage"
	"github.com/prattmic/restic-remote/api"
	"github.com/prattmic/restic-remote/binver"
	"github.com/prattmic/restic-remote/log"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

type updateOpts struct {
	release *api.Release

	updateRestic bool
	updateClient bool

	resticPath string
	clientPath string

	googleCredsFile string

	// binaryBucket is the GCS bucket containing the binaries referenced in
	// release. This is just the raw bucket name, no "gs://".
	binaryBucket string
}

func updateCheck(ctx context.Context, a *api.API) error {
	release, err := a.GetRelease()
	if err != nil {
		return fmt.Errorf("error getting current release: %v", err)
	}

	log.Infof("Latest release: %+v", release)

	opts := updateOpts{
		release: release,
	}

	opts.googleCredsFile = viper.GetString("restic.google-credentials")  // TODO: not restic.
	if opts.googleCredsFile == "" {
		return fmt.Errorf("no Google credentials")
	}

	opts.binaryBucket = viper.GetString("google.binary-bucket")
	if opts.binaryBucket == "" {
		return fmt.Errorf("binary bucket not configured")
	}

	opts.resticPath = viper.GetString("restic.binary")
	if opts.resticPath == "" {
		return fmt.Errorf("restic path unknown")
	}

	opts.clientPath, err = os.Executable()
	if err != nil {
		return fmt.Errorf("error getting client path: %v", err)
	}

	rver, err := binver.Restic(opts.resticPath)
	if err != nil {
		return fmt.Errorf("error getting restic version: %v", err)
	}

	log.Infof("Current restic version: %s", rver)
	log.Infof("Current client version: %s", versionStr)

	opts.updateRestic = rver != release.ResticVersion
	opts.updateClient = versionStr != release.ClientVersion
	if !opts.updateRestic && !opts.updateClient {
		log.Infof("No updates available")
		return nil
	}

	return performUpdate(ctx, opts)
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

// downloadToTmp downloads the new release of the existing binary bin to a
// temporary file in the same folder.
//
// Returns the name of the file.
func downloadToTmp(ctx context.Context, bkt *storage.BucketHandle, release *api.Release, bin string) (string, error) {
	// Name of the binary to grab from the bucket. Differentiates between
	// restic/client and linux/windows (exe).
	base := filepath.Base(bin)
	dir := filepath.Dir(bin)

	// Create tmp file in the same folder so we won't accidentally
	// try to do a cross-mount rename later.
	f, err := tempExecutable(dir, "tmp")
	if err != nil {
		return "", fmt.Errorf("error creating tmpfile: %v", err)
	}
	defer f.Close()
	name := f.Name()

	src := path.Join(release.Path, base)
	if err := download(ctx, f, bkt, src); err != nil {
		return "", fmt.Errorf("error downloading %s: %v", src, err)
	}

	return name, nil
}

func performUpdate(ctx context.Context, opts updateOpts) error {
	resticBin := "restic"
	clientBin := "client"
	if runtime.GOOS == "windows" {
		resticBin += ".exe"
		clientBin += ".exe"
	}

	c, err := storage.NewClient(ctx, option.WithCredentialsFile(opts.googleCredsFile))
	if err != nil {
		return fmt.Errorf("error creating storage client: %v", err)
	}

	bkt := c.Bucket(opts.binaryBucket)

	var tmpRestic, tmpClient string
	if opts.updateRestic {
		tmpRestic, err = downloadToTmp(ctx, bkt, opts.release, opts.resticPath)
		if err != nil {
			return fmt.Errorf("error downloading restic: %v", err)
		}
		defer os.Remove(tmpRestic)
	}
	if opts.updateClient {
		tmpClient, err = downloadToTmp(ctx, bkt, opts.release, opts.clientPath)
		if err != nil {
			return fmt.Errorf("error downloading client: %v", err)
		}
		defer os.Remove(tmpClient)
	}

	return checkAndInstall(opts, tmpRestic, tmpClient)
}

func checkAndInstall(opts updateOpts, tmpRestic, tmpClient string) (err error) {
	// Make sure we got working binaries with the correct versions.

	if opts.updateRestic {
		version, err := binver.Restic(tmpRestic)
		if err != nil {
			return fmt.Errorf("error getting version of new restic %s: %v", tmpRestic, err)
		}
		if version != opts.release.ResticVersion {
			return fmt.Errorf("restic version mismatch got %q want %q", version, opts.release.ResticVersion)
		}
	}
	if opts.updateClient {
		version, err := binver.Client(tmpClient)
		if err != nil {
			return fmt.Errorf("error getting version of new client %s: %v", tmpClient, err)
		}
		if version != opts.release.ClientVersion {
			return fmt.Errorf("client version mismatch got %q want %q", version, opts.release.ClientVersion)
		}
	}

	// Move the existing binaries out of the way (and keep them as
	// backups).
	var move []string
	if opts.updateRestic {
		move = append(move, opts.resticPath)
	}
	if opts.updateClient {
		move = append(move, opts.clientPath)
	}

	for _, p := range move {
		if err := os.Rename(p, p+".old"); err != nil {
			return fmt.Errorf("error moving old binary %s: %v", p, err)
		}
		defer func() {
			if err == nil {
				return
			}

			// Something went wrong, revert.
			log.Errorf("Encountered error: %v; moving old binary back", err)
			if err := os.Rename(p+".old", p); err != nil {
				log.Errorf("Failed to move %s back: %v", p, err)
			}
		}()
	}

	// Finally, move the existing binaries into place.
	if opts.updateRestic {
		if err := os.Rename(tmpRestic, opts.resticPath); err != nil {
			return fmt.Errorf("error moving new restic: %v", err)
		}
	}
	if opts.updateClient {
		if err := os.Rename(tmpClient, opts.clientPath); err != nil {
			return fmt.Errorf("error moving new client: %v", err)
		}
	}

	// Success! Re-exec to the new version. This isn't actually needed if
	// we only updated restic, but it is simple enough to restart.
	log.Infof("Updated, restarting...")
	execve(opts.clientPath, os.Args[1:])

	return nil
}
